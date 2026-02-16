package seed

import (
	"context"

	"github.com/google/uuid"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/infra/repo"
)

type ProductTypeSeed struct {
	repo *repo.ProductTypeRepository
}

func NewProductTypeSeed(repo *repo.ProductTypeRepository) *ProductTypeSeed {
	return &ProductTypeSeed{repo: repo}
}

func (s *ProductTypeSeed) Name() string {
	return "Create ProductTypes"
}

func (s *ProductTypeSeed) Run(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return s.repo.CreateBatch(ctx, ProductTypesEnt())
	}
}

func ProductTypesEnt() []domain.ProductType {
	return []domain.ProductType{
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
