package logger

import (
	"context"
	"log/slog"

	"github.com/VaLeraGav/avito-pvz-service/pkg/request_id"
)

// ctxHandler — обёртка для slog.Handler, добавляющая request_id из контекста
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
	if reqID := request_id.GetReqID(ctx); reqID != "" {
		r.AddAttrs(slog.String(request_id.ReqIDKey, reqID))
	}
	return h.Handler.Handle(ctx, r)
}
