package logger

import (
	"context"
	"log/slog"

	"github.com/valeragav/avito-pvz-service/pkg/requestid"
)

// ctxHandler — обёртка для slog.Handler, добавляющая requestid из контекста
type ctxHandler struct {
	slog.Handler
}

func (h *ctxHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ctxHandler{h.Handler.WithAttrs(attrs)}
}

func (h *ctxHandler) WithGroup(name string) slog.Handler {
	return &ctxHandler{h.Handler.WithGroup(name)}
}

func (h *ctxHandler) Handle(ctx context.Context, r slog.Record) error {
	if reqID := requestid.GetReqID(ctx); reqID != "" {
		r.AddAttrs(slog.String(requestid.LogFieldRequestID, reqID))
	}

	return h.Handler.Handle(ctx, r)
}
