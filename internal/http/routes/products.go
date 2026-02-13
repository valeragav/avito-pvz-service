//nolint:dupl // routes are intentionally separated for clarity
package routes

import (
	"github.com/VaLeraGav/avito-pvz-service/internal/http/handlers/products"
	"github.com/VaLeraGav/avito-pvz-service/internal/http/middleware"
	"github.com/VaLeraGav/avito-pvz-service/internal/security"
	authService "github.com/VaLeraGav/avito-pvz-service/internal/service/auth"
	"github.com/go-chi/chi"
)

type ProductsRoute struct {
	productsHandlers *products.ProductsHandlers
	jwtService       *security.JwtService
}

func NewProductsRoute(productsHandlers *products.ProductsHandlers, jwtService *security.JwtService) *ProductsRoute {
	return &ProductsRoute{
		productsHandlers,
		jwtService,
	}
}

func (router ProductsRoute) Init(r chi.Router) {
	r.Route("/products", func(b chi.Router) {
		b.Use(middleware.AuthMiddleware(router.jwtService))

		b.With(middleware.RequireRoles(authService.EmployeeRole)).Post("/", router.productsHandlers.Create)
	})
}
