package seed

import (
	"github.com/google/uuid"
	"github.com/valeragav/avito-pvz-service/internal/domain"
)

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
