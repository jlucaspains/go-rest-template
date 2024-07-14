package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTraceMiddleware(t *testing.T) {
	// Setup
	r := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Context().Value(ContextKey("traceId")) != nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	})

	// Execution
	TraceMiddleware(next).ServeHTTP(w, r)

	// Validation
	assert.Equal(t, http.StatusOK, w.Code)
}
