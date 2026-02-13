package auth

import (
	"context"

	"github.com/valeragav/avito-pvz-service/internal/infrastructure/storage/user"
)

type RepositoryUser interface {
	Create(ctx context.Context, user user.User) (*user.User, error)
	Get(ctx context.Context, filter user.User) (*user.User, error)
}
