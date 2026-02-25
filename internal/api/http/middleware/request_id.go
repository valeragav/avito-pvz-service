package middleware

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/valeragav/avito-pvz-service/pkg/logger"
	"github.com/valeragav/avito-pvz-service/pkg/requestid"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.NewString()
		ctx := r.Context()

		ctx = requestid.SetReqID(ctx, requestID)

		ctx = logger.WithCtx(ctx, slog.String(requestid.LogFieldRequestID, requestID))

		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
