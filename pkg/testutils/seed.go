package testutils

import (
	"context"

	"github.com/valeragav/avito-pvz-service/internal/infra/postgres"
	"github.com/valeragav/avito-pvz-service/internal/seed"
	"github.com/valeragav/avito-pvz-service/pkg/seeder"
)

type SeedTarget int

const (
	SeedCities SeedTarget = iota
	SeedReceptionStatuses
	SeedProductTypes
)

func Seed(ctx context.Context, qeProvider postgres.QueryEngineProvider, targets ...SeedTarget) error {
	if len(targets) == 0 {
		targets = []SeedTarget{SeedCities, SeedReceptionStatuses, SeedProductTypes}
	}

	sd := seeder.New()

	for _, t := range targets {
		switch t {
		case SeedCities:
			cityRepo := postgres.NewCityRepository(qeProvider)
			sd.Add(seeder.NewGenericSeed("Create Cities", cityRepo, seed.CitiesEnt))
		case SeedReceptionStatuses:
			statusRepo := postgres.NewReceptionStatusRepository(qeProvider)
			sd.Add(seeder.NewGenericSeed("Create ReceptionStatuses", statusRepo, seed.StatusesEnt))
		case SeedProductTypes:
			productTypeRepo := postgres.NewProductTypeRepository(qeProvider)
			sd.Add(seeder.NewGenericSeed("Create ProductTypes", productTypeRepo, seed.ProductTypesEnt))
		}
	}

	return sd.Run(ctx)
}
