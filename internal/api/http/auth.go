package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers/auth"
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
