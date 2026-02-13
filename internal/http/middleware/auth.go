package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/VaLeraGav/avito-pvz-service/internal/http/handlers/response"
	"github.com/VaLeraGav/avito-pvz-service/internal/security"
	"github.com/VaLeraGav/avito-pvz-service/internal/service/auth"
)

const prefixAuth = "Bearer "

type ContextRole struct{}

type Handler func(w http.ResponseWriter, r *http.Request)

type JwtService interface {
	ValidateJwt(string) (*security.Claims, error)
}

func AuthMiddleware(jwtService JwtService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.WriteError(w, ctx, http.StatusUnauthorized, "authorization required", nil)
				return
			}

			jvtToken := strings.TrimPrefix(authHeader, prefixAuth)
			if jvtToken == "" {
				response.WriteError(w, ctx, http.StatusUnauthorized, "invalid token format", nil)
				return
			}

			claims, err := jwtService.ValidateJwt(jvtToken)
			if err != nil {
				response.WriteError(w, ctx, http.StatusUnauthorized, err.Error(), nil)
				return
			}

			ctx = context.WithValue(ctx, ContextRole{}, claims.Role)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

func RequireRoles(roles ...auth.UserRole) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			userRole, ok := ctx.Value(ContextRole{}).(string)

			if !ok {
				response.WriteError(w, ctx, http.StatusUnauthorized, "unauthorized", nil)
				return
			}

			for _, role := range roles {
				if userRole == string(role) {
					next.ServeHTTP(w, r)
					return
				}
			}

			response.WriteError(w, ctx, http.StatusForbidden, "permission denied", nil)
		})
	}
}
