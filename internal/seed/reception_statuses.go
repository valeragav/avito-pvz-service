package seed

import (
	"github.com/google/uuid"
	"github.com/valeragav/avito-pvz-service/internal/domain"
)

func StatusesEnt() []domain.ReceptionStatus {
	return []domain.ReceptionStatus{
		{
			ID:   uuid.New(),
			Name: "in_progress",
		},
		{
			ID:   uuid.New(),
			Name: "close",
		},
	}
}
