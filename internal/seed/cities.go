package seed

import (
	"context"

	"github.com/VaLeraGav/avito-pvz-service/internal/infrastructure/storage/cities"
	"github.com/google/uuid"
)

type CitiesSeed struct {
	repo *cities.Repository
}

func NewCitiesSeed(repo *cities.Repository) *CitiesSeed {
	return &CitiesSeed{repo: repo}
}

func (s *CitiesSeed) Name() string {
	return "Create Cities"
}

func (s *CitiesSeed) Run(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		_, err := s.repo.CreateBatch(ctx, CitiesEnt())
		return err
	}
}

func CitiesEnt() []cities.Cities {
	return []cities.Cities{
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
