package routes

import (
	"github.com/VaLeraGav/avito-pvz-service/internal/http/handlers/auth"
	"github.com/go-chi/chi"
)

type AuthRoute struct {
	authHandlers *auth.AuthHandlers
}

func NewAuthRoute(authHandlers *auth.AuthHandlers) *AuthRoute {
	return &AuthRoute{
		authHandlers: authHandlers,
	}
}

func (router AuthRoute) Init(r chi.Router) {
	r.Post("/dummyLogin", router.authHandlers.DummyLogin)
	r.Post("/register", router.authHandlers.Register)
	r.Post("/login", router.authHandlers.Login)
}
