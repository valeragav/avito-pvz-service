package middleware

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers/response"
	"github.com/valeragav/avito-pvz-service/internal/domain"
)

const prefixAuth = "Bearer "

type ContextRole struct{}

type Handler func(w http.ResponseWriter, r *http.Request)

type JwtService interface {
	ValidateJwt(string) (*domain.UserClaims, error)
}

type AuthMiddleware struct {
	jwtService JwtService
}

func NewAuthMiddleware(jwtService JwtService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService,
	}
}

func (a AuthMiddleware) Init() func(next http.Handler) http.Handler {
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

			claims, err := a.jwtService.ValidateJwt(jvtToken)
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

func (a AuthMiddleware) RequireRoles(roles ...domain.Role) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			userRole, ok := ctx.Value(ContextRole{}).(domain.Role)

			if !ok {
				response.WriteError(w, ctx, http.StatusUnauthorized, "unauthorized", nil)
				return
			}

			if slices.Contains(roles, userRole) {
				next.ServeHTTP(w, r)
				return
			}

			response.WriteError(w, ctx, http.StatusForbidden, "permission denied", nil)
		})
	}
}
