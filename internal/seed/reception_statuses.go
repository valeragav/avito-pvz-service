package seed

import (
	"context"

	"github.com/google/uuid"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/pkg/seeder"
)

type ReceptionStatusSeed struct {
	repo seeder.SeedRepository[domain.ReceptionStatus]
}

func NewReceptionStatusSeed(repo seeder.SeedRepository[domain.ReceptionStatus]) *ReceptionStatusSeed {
	return &ReceptionStatusSeed{repo: repo}
}

func (s *ReceptionStatusSeed) Name() string {
	return "Create Reception Statuses"
}

func (s *ReceptionStatusSeed) Run(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return s.repo.CreateBatch(ctx, StatusesEnt())
	}
}

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
