package seed

import (
	"context"

	"github.com/google/uuid"
	"github.com/valeragav/avito-pvz-service/internal/infrastructure/storage/product_types"
)

type ProductTypesSeed struct {
	repo *product_types.Repository
}

func NewProductTypesSeed(repo *product_types.Repository) *ProductTypesSeed {
	return &ProductTypesSeed{repo: repo}
}

func (s *ProductTypesSeed) Name() string {
	return "Create ProductTypes"
}

func (s *ProductTypesSeed) Run(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		_, err := s.repo.CreateBatch(ctx, ProductTypesEnt())
		return err
	}
}

func ProductTypesEnt() []product_types.ProductTypes {
	return []product_types.ProductTypes{
		{
			ID:   uuid.New(),
			Name: "электроника",
		},
		{
			ID:   uuid.New(),
			Name: "одежда",
		},
		{
			ID:   uuid.New(),
			Name: "обувь",
		},
	}
}
