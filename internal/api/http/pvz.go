package http

import (
	"github.com/go-chi/chi"
	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers/product"
	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers/reception"

	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers/pvz"

	"github.com/valeragav/avito-pvz-service/internal/api/http/middleware"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/security"
)

type PVZRoute struct {
	pvzHandlers        *pvz.PVZHandlers
	receptionsHandlers *reception.ReceptionHandlers
	productsHandlers   *product.ProductHandlers
	jwtService         *security.JwtService
}

func NewPVZRoute(
	pvzHandlers *pvz.PVZHandlers,
	receptionsHandlers *reception.ReceptionHandlers,
	productsHandlers *product.ProductHandlers,
	jwtService *security.JwtService,
) *PVZRoute {
	return &PVZRoute{
		pvzHandlers,
		receptionsHandlers,
		productsHandlers,
		jwtService,
	}
}

func (router PVZRoute) Init(r chi.Router) {
	r.Route("/pvz", func(b chi.Router) {
		b.Use(middleware.AuthMiddleware(router.jwtService))

		b.With(middleware.RequireRoles(domain.EmployeeRole, domain.ModeratorRole)).Get("/", router.pvzHandlers.List)
		b.With(middleware.RequireRoles(domain.ModeratorRole)).Post("/", router.pvzHandlers.Create)

		b.With(middleware.RequireRoles(domain.EmployeeRole)).Post("/{pvzID}/close_last_reception", router.receptionsHandlers.CloseLastReception)
		b.With(middleware.RequireRoles(domain.EmployeeRole)).Post("/{pvzID}/delete_last_product", router.productsHandlers.DeleteLastProduct)
	})
}
