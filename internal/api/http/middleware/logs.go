package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/valeragav/avito-pvz-service/pkg/logger"
	"github.com/valeragav/avito-pvz-service/pkg/request_id"
)

func NewLogger(log *logger.Logger) func(next http.Handler) http.Handler {
	return middleware.RequestLogger(&chiLoggerMiddleware{log})
}

type chiLoggerMiddleware struct {
	log *logger.Logger
}

func (m *chiLoggerMiddleware) NewLogEntry(r *http.Request) middleware.LogEntry {
	log := m.log.With(
		slog.String(request_id.LogFieldRequestID, request_id.GetReqID(r.Context())),
		slog.String("http_method", r.Method),
		slog.String("remote_addr", r.RemoteAddr),
		slog.String("uri", r.RequestURI),
	)

	log.Info("request started")

	return &chiLoggerEntry{log: log}
}

type chiLoggerEntry struct {
	log *logger.Logger
}

func (l *chiLoggerEntry) Write(status, bytes int, _ http.Header, elapsed time.Duration, _ any) {
	l.log.Info("request completed",
		slog.Int("resp_status", status),
		slog.Int("resp_bytes_length", bytes),
		slog.String("resp_elapsed", elapsed.Round(time.Millisecond/100).String()),
	)
}

func (l *chiLoggerEntry) Panic(v any, stack []byte) {
	l.log.Error("panic recovered",
		slog.String("panic", fmt.Sprintf("%v", v)),
		slog.String("stack", string(stack)),
	)
}
