//nolint:dupl // routes are intentionally separated for clarity
package http

import (
	"github.com/go-chi/chi"
	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers/product"
	"github.com/valeragav/avito-pvz-service/internal/api/http/middleware"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/security"
)

type ProductsRoute struct {
	productsHandlers *product.ProductHandlers
	jwtService       *security.JwtService
}

func NewProductsRoute(productsHandlers *product.ProductHandlers, jwtService *security.JwtService) *ProductsRoute {
	return &ProductsRoute{
		productsHandlers,
		jwtService,
	}
}

func (router ProductsRoute) Init(r chi.Router) {
	r.Route("/products", func(b chi.Router) {
		b.Use(middleware.AuthMiddleware(router.jwtService))

		b.With(middleware.RequireRoles(domain.EmployeeRole)).Post("/", router.productsHandlers.Create)
	})
}
