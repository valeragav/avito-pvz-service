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
	"github.com/valeragav/avito-pvz-service/internal/app"
	"github.com/valeragav/avito-pvz-service/pkg/logger"
)

func NewRouter(app *app.App) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middlewareChi.RealIP)

	router.Use(middleware.RequestID)

	// TODO: в config передавать cors options
	router.Use(middleware.Cors(nil))
	router.Use(middleware.NewLogger(logger.GetLogger()))
	router.Use(middlewareChi.Recoverer) // обязательно после NewLogger
	router.Use(middleware.Metrics)

	authMiddleware := middleware.NewAuthMiddleware(app.JwtService)

	router.HandleFunc("/ping", handlers.PingHandler)

	authHandlers := auth.New(app.Validator, app.AuthUseCase)
	pvzHandlers := pvz.New(app.Validator, app.PVZUseCase)
	receptionsHandlers := reception.New(app.Validator, app.ReceptionUseCase)
	productsHandlers := product.New(app.Validator, app.ProductUseCase)

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
