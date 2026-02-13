package receptions

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/VaLeraGav/avito-pvz-service/internal/infrastructure/storage"
	"github.com/VaLeraGav/avito-pvz-service/internal/infrastructure/storage/receptions"
	"github.com/VaLeraGav/avito-pvz-service/internal/infrastructure/storage/statuses"
	"github.com/google/uuid"
)

type ReceptionsService struct {
	receptionsRepo *receptions.Repository
	statusesRepo   *statuses.Repository
}

func New(receptionsRepo *receptions.Repository, statusesRepo *statuses.Repository) *ReceptionsService {
	return &ReceptionsService{
		receptionsRepo,
		statusesRepo,
	}
}

var ErrReceptionsNotClosed = errors.New("previous receptions of the goods was not closed")

func (s *ReceptionsService) Create(ctx context.Context, createIn CreateIn) (*CreateOut, error) {
	const op = "receptions.Create"

	// Если же предыдущая приёмка товара не была закрыта, то операция по созданию нового приёма товаров невозможна.
	receptionLast, err := s.receptionsRepo.GetLastWithStatus(ctx, receptions.Receptions{
		PvzID: createIn.PvzID,
	})
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return nil, fmt.Errorf("%s: failed to check last reception status: %w", op, err)
	}
	if receptionLast != nil && receptionLast.StatusName != statuses.StatusClose {
		return nil, ErrReceptionsNotClosed
	}

	status, err := s.statusesRepo.Get(ctx, statuses.Statuses{
		Name: statuses.StatusInProgress,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get status: %w", op, err)
	}

	pvzRes, err := s.receptionsRepo.Create(ctx, receptions.Receptions{
		DateTime: time.Now(),
		PvzID:    createIn.PvzID,
		StatusID: status.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create reception: %w", op, err)
	}

	return &CreateOut{
		ID:       pvzRes.ID,
		DateTime: pvzRes.DateTime,
		PvzID:    pvzRes.PvzID,
		Status:   status.Name,
	}, nil
}

func (s *ReceptionsService) CloseLastReception(ctx context.Context, pvzID uuid.UUID) (*CloseLastReceptionOut, error) {
	const op = "receptions.CloseLastReception"

	lastReception, err := s.receptionsRepo.GetLastWithStatus(ctx, receptions.Receptions{
		PvzID: pvzID,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to find pvz: %w", op, err)
	}
	if lastReception.StatusName == statuses.StatusClose {
		return nil, ErrNotFoundReceptionsRepoInProgress
	}

	status, err := s.statusesRepo.Get(ctx, statuses.Statuses{
		Name: statuses.StatusClose,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get status: %w", op, err)
	}

	closedReception, err := s.receptionsRepo.Update(ctx, lastReception.ID, receptions.Receptions{
		StatusID: status.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to close reception: %w", err)
	}

	return &CloseLastReceptionOut{
		ID:       closedReception.ID,
		DateTime: closedReception.DateTime,
		PvzID:    closedReception.PvzID,
		Status:   status.Name,
	}, nil
}
