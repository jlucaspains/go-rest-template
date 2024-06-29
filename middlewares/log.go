package middlewares

import (
	"bytes"
	"log/slog"
	"net/http"
	"time"
)

type LogResponseWriter struct {
	http.ResponseWriter
	statusCode int
	buf        bytes.Buffer
}

func (w *LogResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *LogResponseWriter) Write(body []byte) (int, error) {
	w.buf.Write(body)
	return w.ResponseWriter.Write(body)
}

func newLogResponseWriter(w http.ResponseWriter) *LogResponseWriter {
	return &LogResponseWriter{ResponseWriter: w}
}

func LogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		logRespWriter := newLogResponseWriter(w)
		next.ServeHTTP(logRespWriter, r)

		slog.Info(
			"WebRequest",
			"proto", r.Proto,
			"method", r.Method,
			"url", r.URL,
			"duration", time.Since(startTime),
			"status", logRespWriter.statusCode)

	})
}
