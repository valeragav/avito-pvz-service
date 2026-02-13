package receptions

import (
	"context"

	"github.com/google/uuid"
	"github.com/valeragav/avito-pvz-service/internal/service/receptions"
)

//go:generate mockgen -source=contract.go -destination=./mocks/service_mock.go -package=mocks
type receptionsService interface {
	CloseLastReception(ctx context.Context, pvzID uuid.UUID) (*receptions.CloseLastReceptionOut, error)
	Create(ctx context.Context, createIn receptions.CreateIn) (*receptions.CreateOut, error)
}
