package middlewares

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogMiddleware(t *testing.T) {
	var buffer *bytes.Buffer = new(bytes.Buffer)
	logger := slog.New(slog.NewTextHandler(buffer, nil))
	slog.SetDefault(logger)

	// Setup
	r := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Execution
	LogMiddleware(next).ServeHTTP(w, r)

	t.Log(buffer.String())

	// Validation
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Regexp(t, "time=[0-9T\\-:\\.]+ level=INFO msg=WebRequest proto=HTTP/1.1 method=GET url=/test duration=0s status=200", buffer.String())
}
