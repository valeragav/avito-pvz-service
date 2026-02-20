//nolint:dupl // routes are intentionally separated for clarity
package http

import (
	"github.com/go-chi/chi"
	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers/reception"
	"github.com/valeragav/avito-pvz-service/internal/api/http/middleware"
	"github.com/valeragav/avito-pvz-service/internal/domain"
)

type ReceptionsRoute struct {
	authMiddleware     *middleware.AuthMiddleware
	receptionsHandlers *reception.ReceptionHandlers
}

func NewReceptionsRoute(authMiddleware *middleware.AuthMiddleware, receptionsHandlers *reception.ReceptionHandlers) *ReceptionsRoute {
	return &ReceptionsRoute{
		authMiddleware,
		receptionsHandlers,
	}
}

func (router ReceptionsRoute) Init(r chi.Router) {
	r.Route("/receptions", func(b chi.Router) {
		b.Use(router.authMiddleware.Init())

		b.With(router.authMiddleware.RequireRoles(domain.EmployeeRole)).Post("/", router.receptionsHandlers.Create)
	})
}
