package auth

import (
	"github.com/VaLeraGav/avito-pvz-service/internal/service/auth"
	"github.com/google/uuid"
)

type DummyLoginRequest struct {
	Role string `json:"role" validate:"required,oneofci=employee moderator"`
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=255"`
	Role     string `json:"role" validate:"required,oneofci=employee moderator"`
}

type RegisterResponse struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Role  string    `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

func ToRegisterIn(req RegisterRequest) auth.RegisterIn {
	return auth.RegisterIn{
		Email:    req.Email,
		Password: req.Password,
		Role:     auth.UserRole(req.Role),
	}
}

func ToLoginIn(req LoginRequest) auth.LoginIn {
	return auth.LoginIn{
		Email:    req.Email,
		Password: req.Password,
	}
}

func ToRegisterResponse(out auth.RegisterOut) RegisterResponse {
	return RegisterResponse{
		ID:    out.User.ID,
		Email: out.User.Email,
		Role:  out.User.Role,
	}
}
