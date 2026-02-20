//nolint:dupl // routes are intentionally separated for clarity
package http

import (
	"github.com/go-chi/chi"
	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers/product"
	"github.com/valeragav/avito-pvz-service/internal/api/http/middleware"
	"github.com/valeragav/avito-pvz-service/internal/domain"
)

type ProductsRoute struct {
	authMiddleware   *middleware.AuthMiddleware
	productsHandlers *product.ProductHandlers
}

func NewProductsRoute(authMiddleware *middleware.AuthMiddleware, productsHandlers *product.ProductHandlers) *ProductsRoute {
	return &ProductsRoute{
		authMiddleware,
		productsHandlers,
	}
}

func (router ProductsRoute) Init(r chi.Router) {
	r.Route("/products", func(b chi.Router) {
		b.Use(router.authMiddleware.Init())

		b.With(router.authMiddleware.RequireRoles(domain.EmployeeRole)).Post("/", router.productsHandlers.Create)
	})
}
