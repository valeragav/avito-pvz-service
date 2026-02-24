package http

import (
	"github.com/go-chi/chi/v5"
	middlewareChi "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/valeragav/avito-pvz-service/api/v1/swagger"
	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers"
	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers/auth"
	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers/product"
	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers/pvz"
	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers/reception"
	"github.com/valeragav/avito-pvz-service/internal/api/http/middleware"
	"github.com/valeragav/avito-pvz-service/internal/app"
	"github.com/valeragav/avito-pvz-service/internal/config"
	"github.com/valeragav/avito-pvz-service/pkg/logger"
)

func NewRouter(appService *app.App, cfg *config.Config) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middlewareChi.RealIP)

	router.Use(middleware.RequestID)

	router.Use(middleware.Concurrency(cfg.HTTPServer.MaxConcurrentRequests))

	router.Use(middleware.Cors(nil))
	router.Use(middleware.NewLogger(logger.GetLogger()))
	router.Use(middlewareChi.Recoverer) // обязательно после NewLogger
	router.Use(middleware.Metrics)

	authMiddleware := middleware.NewAuthMiddleware(appService.JwtService)

	router.HandleFunc("/ping", handlers.PingHandler)

	authHandlers := auth.New(appService.Validator, appService.AuthUseCase)
	pvzHandlers := pvz.New(appService.Validator, appService.PVZUseCase)
	receptionsHandlers := reception.New(appService.Validator, appService.ReceptionUseCase)
	productsHandlers := product.New(appService.Validator, appService.ProductUseCase)

	authRoute := NewAuthRoute(authHandlers)
	authRoute.Init(router)

	pvzRoute := NewPVZRoute(authMiddleware, pvzHandlers, receptionsHandlers, productsHandlers)
	pvzRoute.Init(router)

	productsRoute := NewProductsRoute(authMiddleware, productsHandlers)
	productsRoute.Init(router)

	receptionsRoute := NewReceptionsRoute(authMiddleware, receptionsHandlers)
	receptionsRoute.Init(router)

	return router
}

func NewMetricsRoute() *chi.Mux {
	router := chi.NewRouter()
	router.Handle("/metrics", promhttp.Handler())
	return router
}

func NewSwaggerRoute() *chi.Mux {
	router := chi.NewRouter()
	router.Handle("/swagger/*", httpSwagger.Handler())
	return router
}
