package reception

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/infra"
	"github.com/valeragav/avito-pvz-service/internal/usecase/dto"
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

type transactionManager interface {
	RunRepeatableRead(ctx context.Context, fn func(ctx context.Context) error) error
	RunReadCommitted(ctx context.Context, fn func(ctx context.Context) error) error
}

type ReceptionUseCase struct {
	tm            transactionManager
	receptionRepo receptionRepo
	statusRepo    receptionStatusRepo
	pvzRepo       pvzRepo
}

func New(tm transactionManager, receptionRepo receptionRepo, statusRepo receptionStatusRepo, pvzRepo pvzRepo) *ReceptionUseCase {
	return &ReceptionUseCase{
		tm,
		receptionRepo,
		statusRepo,
		pvzRepo,
	}
}

func (s *ReceptionUseCase) Create(ctx context.Context, createIn dto.ReceptionCreate) (*domain.Reception, error) {
	const op = "receptions.Create"

	if err := s.checkPVZExists(ctx, createIn.PvzID); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var result *domain.Reception

	err := s.tm.RunRepeatableRead(ctx, func(ctx context.Context) error {
		var txErr error
		result, txErr = s.create(ctx, createIn.PvzID)
		return txErr
	})

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return result, nil
}

func (s *ReceptionUseCase) CloseLastReception(ctx context.Context, pvzID uuid.UUID) (*domain.Reception, error) {
	const op = "receptions.CloseLastReception"

	if err := s.checkPVZExists(ctx, pvzID); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var result *domain.Reception

	err := s.tm.RunReadCommitted(ctx, func(ctx context.Context) error {
		var txErr error
		result, txErr = s.closeLastReception(ctx, pvzID)
		return txErr
	})

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return result, nil
}

func (s *ReceptionUseCase) create(ctx context.Context, pvzID uuid.UUID) (*domain.Reception, error) {
	// Если же предыдущая приёмка товара не была закрыта, то операция по созданию нового приёма товаров невозможна.
	_, err := s.receptionRepo.FindByStatus(ctx, domain.ReceptionStatusInProgress, domain.Reception{
		PvzID: pvzID,
	})
	if err == nil {
		return nil, domain.ErrNoReceptionIsCurrentlyInProgress
	}
	if !errors.Is(err, infra.ErrNotFound) {
		return nil, fmt.Errorf("failed to check last reception status: %w", err)
	}

	status, err := s.statusRepo.Get(ctx, domain.ReceptionStatus{
		Name: domain.ReceptionStatusInProgress,
	})
	if err != nil {
		if errors.Is(err, infra.ErrNotFound) {
			return nil, domain.ErrStatusNotFound
		}
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	pvzRes, err := s.receptionRepo.Create(ctx, domain.Reception{
		DateTime: time.Now(),
		PvzID:    pvzID,
		StatusID: status.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create reception: %w", err)
	}
	pvzRes.ReceptionStatus = status
	return pvzRes, nil
}

func (s *ReceptionUseCase) closeLastReception(ctx context.Context, pvzID uuid.UUID) (*domain.Reception, error) {
	lastReception, err := s.receptionRepo.FindByStatus(ctx, domain.ReceptionStatusInProgress, domain.Reception{
		PvzID: pvzID,
	})
	if err != nil {
		if errors.Is(err, infra.ErrNotFound) {
			return nil, domain.ErrReceptionNotFound
		}
		return nil, fmt.Errorf("failed to find reception: %w", err)
	}

	status, err := s.statusRepo.Get(ctx, domain.ReceptionStatus{
		Name: domain.ReceptionStatusClose,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
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

func (s *ReceptionUseCase) checkPVZExists(ctx context.Context, pvzID uuid.UUID) error {
	_, err := s.pvzRepo.Get(ctx, domain.PVZ{ID: pvzID})
	if err != nil {
		if errors.Is(err, infra.ErrNotFound) {
			return domain.ErrPVZNotFound
		}
		return fmt.Errorf("failed to find pvz: %w", err)
	}
	return nil
}
