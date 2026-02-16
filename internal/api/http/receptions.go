//nolint:dupl // routes are intentionally separated for clarity
package http

import (
	"github.com/go-chi/chi"
	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers/reception"
	"github.com/valeragav/avito-pvz-service/internal/api/http/middleware"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/security"
)

type ReceptionsRoute struct {
	receptionsHandlers *reception.ReceptionHandlers
	jwtService         *security.JwtService
}

func NewReceptionsRoute(receptionsHandlers *reception.ReceptionHandlers, jwtService *security.JwtService) *ReceptionsRoute {
	return &ReceptionsRoute{
		receptionsHandlers,
		jwtService,
	}
}

func (router ReceptionsRoute) Init(r chi.Router) {
	r.Route("/receptions", func(b chi.Router) {
		b.Use(middleware.AuthMiddleware(router.jwtService))

		b.With(middleware.RequireRoles(domain.EmployeeRole)).Post("/", router.receptionsHandlers.Create)
	})
}
