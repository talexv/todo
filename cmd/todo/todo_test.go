package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/talexv/todo/internal/task"
)

func initTodoAppTest(t *testing.T) {
	go func() {
		err := run()
		assert.NoError(t, err)
	}()

	time.Sleep(3 * time.Second)
}

func TestTodoE2E(t *testing.T) {
	initTodoAppTest(t)

	urlSrv := "http://localhost:8081"
	client := &http.Client{}

	// Get task
	res, err := http.Get(urlSrv + "/tasks")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)

	var tasks []task.Task

	err = json.NewDecoder(res.Body).Decode(&tasks)
	require.NoError(t, err)
	require.Empty(t, tasks)
	res.Body.Close()

	// Create task
	payload := map[string]string{"title": "New task"}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	res, err = http.Post(urlSrv+"/create", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, res.StatusCode)

	var createdTask *task.Task

	err = json.NewDecoder(res.Body).Decode(&createdTask)
	require.NoError(t, err)
	require.Equal(t, payload["title"], createdTask.Title)
	require.False(t, createdTask.Done)
	res.Body.Close()

	// Update status task
	id := strconv.Itoa(int(createdTask.ID))
	req, err := http.NewRequest(http.MethodPatch, urlSrv+"/tasks/"+id+"/done", nil)
	require.NoError(t, err)

	res, err = client.Do(req)
	require.NoError(t, err)

	var updatedTask *task.Task

	err = json.NewDecoder(res.Body).Decode(&updatedTask)
	require.NoError(t, err)
	require.Equal(t, createdTask.ID, updatedTask.ID)
	require.Equal(t, payload["title"], updatedTask.Title)
	require.True(t, updatedTask.Done)
	res.Body.Close()

	// Delete task
	req, err = http.NewRequest(http.MethodDelete, urlSrv+"/tasks/"+id+"/delete", nil)
	require.NoError(t, err)

	res, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, res.StatusCode)
	res.Body.Close()

	res, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, res.StatusCode)
	res.Body.Close()
}
