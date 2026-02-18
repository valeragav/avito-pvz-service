package http

import (
	"github.com/go-chi/chi"
	middlewareChi "github.com/go-chi/chi/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/valeragav/avito-pvz-service/docs"

	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers"
	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers/auth"
	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers/product"
	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers/pvz"
	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers/reception"
	"github.com/valeragav/avito-pvz-service/internal/api/http/middleware"
	"github.com/valeragav/avito-pvz-service/internal/container"
	"github.com/valeragav/avito-pvz-service/pkg/logger"
)

func NewRouter(cnt *container.DIContainer) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middlewareChi.RealIP)

	router.Use(middleware.RequestID)
	router.Use(middleware.Cors(nil))
	router.Use(middleware.NewLogger(logger.GetLogger()))
	router.Use(middlewareChi.Recoverer) // обязательно после NewLogger
	router.Use(middleware.Metrics)

	router.Handle("/metrics", promhttp.Handler())
	router.HandleFunc("/ping", handlers.PingHandler)
	router.Handle("/swagger/*", httpSwagger.Handler())

	authHandlers := auth.New(cnt.Validator, cnt.AuthUseCase)
	pvzHandlers := pvz.New(cnt.Validator, cnt.PVZUseCase)
	receptionsHandlers := reception.New(cnt.Validator, cnt.ReceptionUseCase)
	productsHandlers := product.New(cnt.Validator, cnt.ProductUseCase)

	authRoute := NewAuthRoute(authHandlers)
	authRoute.Init(router)

	pvzRoute := NewPVZRoute(pvzHandlers, receptionsHandlers, productsHandlers, cnt.JwtService)
	pvzRoute.Init(router)

	productsRoute := NewProductsRoute(productsHandlers, cnt.JwtService)
	productsRoute.Init(router)

	receptionsRoute := NewReceptionsRoute(receptionsHandlers, cnt.JwtService)
	receptionsRoute.Init(router)

	return router
}
