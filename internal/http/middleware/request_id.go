package middleware

import (
	"log/slog"
	"net/http"

	"github.com/VaLeraGav/avito-pvz-service/pkg/logger"
	"github.com/VaLeraGav/avito-pvz-service/pkg/request_id"
	"github.com/google/uuid"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.NewString()
		ctx := r.Context()

		ctx = request_id.SetReqID(ctx, requestID)

		ctx = logger.WithCtx(ctx, slog.String(request_id.LogFieldRequestID, requestID))

		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
