package http

import (
	"github.com/go-chi/chi"
	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers/product"
	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers/reception"

	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers/pvz"

	"github.com/valeragav/avito-pvz-service/internal/api/http/middleware"
	"github.com/valeragav/avito-pvz-service/internal/domain"
)

type PVZRoute struct {
	authMiddleware     *middleware.AuthMiddleware
	pvzHandlers        *pvz.PVZHandlers
	receptionsHandlers *reception.ReceptionHandlers
	productsHandlers   *product.ProductHandlers
}

func NewPVZRoute(
	authMiddleware *middleware.AuthMiddleware,
	pvzHandlers *pvz.PVZHandlers,
	receptionsHandlers *reception.ReceptionHandlers,
	productsHandlers *product.ProductHandlers,
) *PVZRoute {
	return &PVZRoute{
		authMiddleware,
		pvzHandlers,
		receptionsHandlers,
		productsHandlers,
	}
}

func (router PVZRoute) Init(r chi.Router) {
	r.Route("/pvz", func(b chi.Router) {
		b.Use(router.authMiddleware.Init())

		b.With(router.authMiddleware.RequireRoles(domain.EmployeeRole, domain.ModeratorRole)).Get("/", router.pvzHandlers.List)
		b.With(router.authMiddleware.RequireRoles(domain.ModeratorRole)).Post("/", router.pvzHandlers.Create)

		b.With(router.authMiddleware.RequireRoles(domain.EmployeeRole)).Post("/{pvzID}/close_last_reception", router.receptionsHandlers.CloseLastReception)
		b.With(router.authMiddleware.RequireRoles(domain.EmployeeRole)).Post("/{pvzID}/delete_last_product", router.productsHandlers.DeleteLastProduct)
	})
}
