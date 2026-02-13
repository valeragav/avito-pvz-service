package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/valeragav/avito-pvz-service/internal/metrics"
)

func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		rw := &responseWriter{
			statusCode:     http.StatusOK,
			ResponseWriter: w,
		}

		next.ServeHTTP(rw, r)

		duration := time.Since(startTime)

		path := chi.RouteContext(r.Context()).RoutePattern()
		if path == "" {
			path = r.URL.Path
		}

		method := r.Method

		metrics.RestRequestInc(method, path)
		metrics.RestRequestDurationObserve(method, path, duration)
		metrics.RestResponseInc(method, path, rw.statusCode)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
