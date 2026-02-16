package seed

import (
	"context"

	"github.com/google/uuid"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/infra/repo"
)

type CitySeed struct {
	repo *repo.CityRepository
}

func NewCitySeed(repo *repo.CityRepository) *CitySeed {
	return &CitySeed{repo: repo}
}

func (s *CitySeed) Name() string {
	return "Create Cities"
}

func (s *CitySeed) Run(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return s.repo.CreateBatch(ctx, CitiesEnt())
	}
}

func CitiesEnt() []domain.City {
	return []domain.City{
		{
			ID:   uuid.New(),
			Name: "Казань",
		},
		{
			ID:   uuid.New(),
			Name: "Москва",
		},
		{
			ID:   uuid.New(),
			Name: "Санкт-Петербург",
		},
	}
}
