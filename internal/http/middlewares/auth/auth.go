package auth

import (
	"net/http"
	"strings"

	"github.com/VaLeraGav/avito-pvz-service/pkg/logger"
	"github.com/go-chi/chi/middleware"
)

var Unauthorized = "Unauthorized"

func BearerAuth(BearerToken string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logger.Debug(Unauthorized, "header", authHeader)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				logger.Debug(Unauthorized, "header", authHeader)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			token := parts[1]
			if token != BearerToken {
				logger.Debug(Unauthorized, "header", authHeader)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(ww, r)
		})
	}
}
