package products

import (
	"context"

	"github.com/VaLeraGav/avito-pvz-service/internal/service/products"
	"github.com/google/uuid"
)

//go:generate mockgen -source=contract.go -destination=./mocks/service_mock.go -package=mocks
type productsService interface {
	Create(ctx context.Context, createIn products.CreateIn) (*products.CreateOut, error)
	DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error
}
