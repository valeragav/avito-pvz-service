package reception

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/dto"
	"github.com/valeragav/avito-pvz-service/internal/infra"
)

//go:generate ${LOCAL_BIN}/mockgen -source=reception.go -destination=./mocks/reception_mock.go -package=mocks
type receptionRepo interface {
	FindByStatus(ctx context.Context, statusName domain.ReceptionStatusCode, filter domain.Reception) (*domain.Reception, error)
	Create(ctx context.Context, reception domain.Reception) (*domain.Reception, error)
	Update(ctx context.Context, receptionID uuid.UUID, update domain.Reception) (*domain.Reception, error)
}

type receptionStatusRepo interface {
	Get(ctx context.Context, filter domain.ReceptionStatus) (*domain.ReceptionStatus, error)
}

type pvzRepo interface {
	Get(ctx context.Context, filter domain.PVZ) (*domain.PVZ, error)
}

type ReceptionUseCase struct {
	receptionRepo receptionRepo
	statusRepo    receptionStatusRepo
	pvzRepo       pvzRepo
}

func New(receptionRepo receptionRepo, statusRepo receptionStatusRepo, pvzRepo pvzRepo) *ReceptionUseCase {
	return &ReceptionUseCase{
		receptionRepo,
		statusRepo,
		pvzRepo,
	}
}

func (s *ReceptionUseCase) Create(ctx context.Context, createIn dto.ReceptionCreate) (*domain.Reception, error) {
	const op = "receptions.Create"

	// Если же предыдущая приёмка товара не была закрыта, то операция по созданию нового приёма товаров невозможна.
	_, err := s.receptionRepo.FindByStatus(ctx, domain.ReceptionStatusInProgress, domain.Reception{
		PvzID: createIn.PvzID,
	})
	if err == nil {
		return nil, domain.ErrNoReceptionIsCurrentlyInProgress
	}
	if !errors.Is(err, infra.ErrNotFound) {
		return nil, fmt.Errorf("%s: failed to check last reception status: %w", op, err)
	}

	status, err := s.statusRepo.Get(ctx, domain.ReceptionStatus{
		Name: domain.ReceptionStatusInProgress,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get status: %w", op, err)
	}

	pvzRes, err := s.receptionRepo.Create(ctx, domain.Reception{
		DateTime: time.Now(),
		PvzID:    createIn.PvzID,
		StatusID: status.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create reception: %w", op, err)
	}

	pvzRes.ReceptionStatus = status

	return pvzRes, nil
}

func (s *ReceptionUseCase) CloseLastReception(ctx context.Context, pvzID uuid.UUID) (*domain.Reception, error) {
	const op = "receptions.CloseLastReception"

	_, err := s.pvzRepo.Get(ctx, domain.PVZ{
		ID: pvzID,
	})
	if err != nil {
		if errors.Is(err, infra.ErrNotFound) {
			return nil, domain.ErrPVZNotFound
		}
		return nil, fmt.Errorf("%s: failed to find pvz: %w", op, err)
	}

	lastReception, err := s.receptionRepo.FindByStatus(ctx, domain.ReceptionStatusInProgress, domain.Reception{
		PvzID: pvzID,
	})
	if err != nil {
		if errors.Is(err, infra.ErrNotFound) {
			return nil, domain.ErrReceptionNotFound
		}
		return nil, fmt.Errorf("%s: failed to find pvz: %w", op, err)
	}

	status, err := s.statusRepo.Get(ctx, domain.ReceptionStatus{
		Name: domain.ReceptionStatusClose,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get status: %w", op, err)
	}

	closedReception, err := s.receptionRepo.Update(ctx, lastReception.ID, domain.Reception{
		StatusID: status.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to close reception: %w", err)
	}

	closedReception.ReceptionStatus = status

	return closedReception, nil
}
