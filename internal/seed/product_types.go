package seed

import (
	"github.com/google/uuid"
	"github.com/valeragav/avito-pvz-service/internal/domain"
)

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
