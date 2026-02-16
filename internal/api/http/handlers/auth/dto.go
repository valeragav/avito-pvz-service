package auth

import (
	"github.com/google/uuid"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/dto"
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

func ToRegisterIn(req RegisterRequest) dto.RegisterIn {
	return dto.RegisterIn{
		Email:    req.Email,
		Password: req.Password,
		Role:     domain.Role(req.Role),
	}
}

func ToLoginIn(req LoginRequest) dto.LoginIn {
	return dto.LoginIn{
		Email:    req.Email,
		Password: req.Password,
	}
}

func ToRegisterResponse(out domain.User) RegisterResponse {
	return RegisterResponse{
		ID:    out.ID,
		Email: out.Email,
		Role:  string(out.Role),
	}
}
