package seed

import (
	"context"

	"github.com/google/uuid"
	"github.com/valeragav/avito-pvz-service/internal/infrastructure/storage/statuses"
)

type StatusesSeed struct {
	repo *statuses.Repository
}

func NewStatusesSeed(repo *statuses.Repository) *StatusesSeed {
	return &StatusesSeed{repo: repo}
}

func (s *StatusesSeed) Name() string {
	return "Create Statuses"
}

func (s *StatusesSeed) Run(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		_, err := s.repo.CreateBatch(ctx, StatusesEnt())
		return err
	}
}

func StatusesEnt() []statuses.Statuses {
	return []statuses.Statuses{
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
