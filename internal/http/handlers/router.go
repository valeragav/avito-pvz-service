package handlers

import (
	"github.com/VaLeraGav/avito-pvz-service/internal/container"
	"github.com/VaLeraGav/avito-pvz-service/internal/http/middlewares/cors"
	"github.com/VaLeraGav/avito-pvz-service/internal/http/middlewares/request_id"
	"github.com/go-chi/chi"
)

func NewRouter(container *container.DIContainer) *chi.Mux {
	router := chi.NewRouter()

	router.Use(request_id.RequestID)
	router.Use(cors.Handler)

	router.HandleFunc("/ping", PingHandler)

	return router
}
