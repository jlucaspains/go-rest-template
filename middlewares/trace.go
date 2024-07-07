package middlewares

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ContextKey string

func TraceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceId := uuid.New()

		ctx := context.WithValue(r.Context(), ContextKey("traceId"), traceId)

		newR := r.WithContext(ctx)

		next.ServeHTTP(w, newR)
	})
}
