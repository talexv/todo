package middleware_test

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/talexv/todo/internal/middleware"
)

//nolint:gochecknoglobals // TestMain cannot directly pass values to tests
var envFilePath string

func TestMain(m *testing.M) {
	flag.StringVar(&envFilePath, "env", "./.env", "путь до файла .env")
	flag.Parse()

	exitcode := m.Run()
	os.Exit(exitcode)
}

func TestRecoverer(t *testing.T) {
	handler := middleware.Recoverer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("panic")
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	res := w.Result()

	defer res.Body.Close()
	require.Equal(t, http.StatusInternalServerError, res.StatusCode)
}
