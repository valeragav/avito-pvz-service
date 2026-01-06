package request_id

import (
	"net/http"

	"github.com/VaLeraGav/avito-pvz-service/pkg/request_id"
	"github.com/google/uuid"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()

		ctx := request_id.SetReqID(r.Context(), requestID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
