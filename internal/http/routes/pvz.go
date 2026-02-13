package routes

import (
	"github.com/valeragav/avito-pvz-service/internal/http/handlers/products"
	"github.com/valeragav/avito-pvz-service/internal/http/handlers/pvz"
	"github.com/valeragav/avito-pvz-service/internal/http/handlers/receptions"
	"github.com/valeragav/avito-pvz-service/internal/http/middleware"
	authService "github.com/valeragav/avito-pvz-service/internal/service/auth"

	"github.com/go-chi/chi"
	"github.com/valeragav/avito-pvz-service/internal/security"
)

type PvzRoute struct {
	pvzHandlers        *pvz.PvzHandlers
	receptionsHandlers *receptions.ReceptionsHandlers
	productsHandlers   *products.ProductsHandlers
	jwtService         *security.JwtService
}

func NewPvzRoute(
	pvzHandlers *pvz.PvzHandlers,
	receptionsHandlers *receptions.ReceptionsHandlers,
	productsHandlers *products.ProductsHandlers,
	jwtService *security.JwtService,
) *PvzRoute {
	return &PvzRoute{
		pvzHandlers,
		receptionsHandlers,
		productsHandlers,
		jwtService,
	}
}

func (router PvzRoute) Init(r chi.Router) {
	r.Route("/pvz", func(b chi.Router) {
		b.Use(middleware.AuthMiddleware(router.jwtService))

		b.With(middleware.RequireRoles(authService.EmployeeRole, authService.EmployeeRole)).Get("/", router.pvzHandlers.List)
		b.With(middleware.RequireRoles(authService.ModeratorRole)).Post("/", router.pvzHandlers.Create)

		b.With(middleware.RequireRoles(authService.EmployeeRole)).Post("/{pvzID}/close_last_reception", router.receptionsHandlers.CloseLastReception)
		b.With(middleware.RequireRoles(authService.EmployeeRole)).Post("/{pvzID}/delete_last_product", router.productsHandlers.DeleteLastProduct)
	})
}
