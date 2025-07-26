package task_test

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"github.com/talexv/todo/internal/middleware"
	"github.com/talexv/todo/internal/task"
)

//nolint:gochecknoglobals // TestMain cannot directly pass values to tests
var envFilePath string

func TestMain(m *testing.M) {
	flag.StringVar(&envFilePath, "env", "./.env", "путь до файла .env")
	flag.Parse()

	exitcode := m.Run()
	os.Exit(exitcode)
}

func initTestEnv(t *testing.T) string {
	t.Helper()

	if err := godotenv.Load(envFilePath); err != nil {
		t.Logf("файл .env.test не найден: %v", err)
	}

	connString := os.Getenv("DATABASE_URL")
	require.NotEmpty(t, connString)

	return connString
}

func clearTestDB(t *testing.T) {
	t.Helper()

	connString := initTestEnv(t)
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connString)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `TRUNCATE tasks RESTART IDENTITY`)
	require.NoError(t, err)
	pool.Close()
}

func initTestDB(t *testing.T) *task.DB {
	t.Helper()

	connString := initTestEnv(t)
	db, err := task.NewDB(connString)
	require.NoError(t, err)

	return db
}

func initTestHandler(t *testing.T) (http.Handler, *task.DB) {
	t.Helper()

	db := initTestDB(t)
	router := http.NewServeMux()
	task.NewHandler(router, db)
	finalHandler := middleware.Recoverer(middleware.Logging(router))

	t.Cleanup(func() {
		db.Close()
	})

	return finalHandler, db
}

func TestCreateTask(t *testing.T) {
	handler, _ := initTestHandler(t)

	payload := map[string]string{"title": "New Task"}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	res := w.Result()

	defer res.Body.Close()
	require.Equal(t, http.StatusCreated, res.StatusCode)

	var createdTask *task.Task
	err = json.NewDecoder(res.Body).Decode(&createdTask)
	require.NoError(t, err)

	require.Equal(t, payload["title"], createdTask.Title)
	require.False(t, createdTask.Done)

	t.Cleanup(func() {
		clearTestDB(t)
	})
}

func TestUpdateStatusTask(t *testing.T) {
	handler, db := initTestHandler(t)
	newTask, err := db.InsertTask(context.Background(), "Test task")
	require.NoError(t, err)

	id := strconv.Itoa(int(newTask.ID))
	req := httptest.NewRequest(http.MethodPatch, "/tasks/"+id+"/done", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	res := w.Result()

	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)

	var updatedTask *task.Task

	err = json.NewDecoder(res.Body).Decode(&updatedTask)
	require.NoError(t, err)
	require.Equal(t, newTask.ID, updatedTask.ID)
	require.Equal(t, newTask.Title, updatedTask.Title)
	require.True(t, updatedTask.Done)

	t.Cleanup(func() {
		clearTestDB(t)
	})
}

func TestDeleteTask(t *testing.T) {
	handler, db := initTestHandler(t)
	newTask, err := db.InsertTask(context.Background(), "New test task")
	require.NoError(t, err)

	id := strconv.Itoa(int(newTask.ID))
	req := httptest.NewRequest(http.MethodDelete, "/tasks/"+id+"/delete", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	res := w.Result()

	defer res.Body.Close()
	require.Equal(t, http.StatusNoContent, res.StatusCode)

	t.Cleanup(func() {
		clearTestDB(t)
	})
}

//nolint:tparallel // disabled to avoid conflicts with other tests
func TestGetTaskParallel(t *testing.T) {
	handler, _ := initTestHandler(t)

	for i := range 100 {
		t.Run(fmt.Sprintf("subtest-%d", i), func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			res := w.Result()

			defer res.Body.Close()
			require.Equal(t, http.StatusOK, res.StatusCode)
		})
	}
}
