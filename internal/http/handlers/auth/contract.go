package auth

import (
	"context"

	"github.com/VaLeraGav/avito-pvz-service/internal/service/auth"
)

//go:generate mockgen -source=contract.go -destination=./mocks/service_mock.go -package=mocks
type authService interface {
	Register(ctx context.Context, in auth.RegisterIn) (*auth.RegisterOut, error)
	Login(ctx context.Context, in auth.LoginIn) (string, error)
	GenerateToken(role string) (string, error)
}
