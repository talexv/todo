package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/talexv/todo/internal/middleware"
)

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
