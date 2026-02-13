//nolint:dupl // routes are intentionally separated for clarity
package routes

import (
	"github.com/VaLeraGav/avito-pvz-service/internal/http/handlers/receptions"
	"github.com/VaLeraGav/avito-pvz-service/internal/http/middleware"
	"github.com/VaLeraGav/avito-pvz-service/internal/security"
	authService "github.com/VaLeraGav/avito-pvz-service/internal/service/auth"
	"github.com/go-chi/chi"
)

type ReceptionsRoute struct {
	receptionsHandlers *receptions.ReceptionsHandlers
	jwtService         *security.JwtService
}

func NewReceptionsRoute(receptionsHandlers *receptions.ReceptionsHandlers, jwtService *security.JwtService) *ReceptionsRoute {
	return &ReceptionsRoute{
		receptionsHandlers,
		jwtService,
	}
}

func (router ReceptionsRoute) Init(r chi.Router) {
	r.Route("/receptions", func(b chi.Router) {
		b.Use(middleware.AuthMiddleware(router.jwtService))

		b.With(middleware.RequireRoles(authService.EmployeeRole)).Post("/", router.receptionsHandlers.Create)
	})
}
