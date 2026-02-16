package pvz

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/dto"
	"github.com/valeragav/avito-pvz-service/internal/infra"
	"github.com/valeragav/avito-pvz-service/pkg/listparams"
)

//go:generate mockgen -source=pvz.go -destination=./mocks/pvz_mocks.go -package=mocks
type pvzRepo interface {
	Create(ctx context.Context, pvz domain.PVZ) (*domain.PVZ, error)
	ListPvzByAcceptanceDateAndCity(ctx context.Context, pagination listparams.Pagination, startDate *time.Time, endDate *time.Time) ([]*domain.PVZ, error)
}

type cityRepo interface {
	Get(ctx context.Context, filter domain.City) (*domain.City, error)
}

type receptionRepo interface {
	ListByIDsWithStatus(ctx context.Context, receptionIDs []uuid.UUID) ([]*domain.Reception, error)
}

type productRepo interface {
	ListByReceptionIDsWithTypeName(ctx context.Context, receptionIDs []uuid.UUID) ([]*domain.Product, error)
}

type PVZUseCase struct {
	pvzRepo       pvzRepo
	cityRepo      cityRepo
	receptionRepo receptionRepo
	productRepo   productRepo
}

func New(pvzRepo pvzRepo, cityRepo cityRepo, receptionRepo receptionRepo, productRepo productRepo) *PVZUseCase {
	return &PVZUseCase{
		pvzRepo,
		cityRepo,
		receptionRepo,
		productRepo,
	}
}

func (s *PVZUseCase) Create(ctx context.Context, createIn dto.PVZCreate) (*domain.PVZ, error) {
	const op = "pvz.Create"

	city, err := s.cityRepo.Get(ctx, domain.City{
		Name: createIn.CityName,
	})
	if err != nil {
		if errors.Is(err, infra.ErrNotFound) {
			return nil, domain.ErrCityNotFound
		}
		return nil, fmt.Errorf("%s: failed to get city: %w", op, err)
	}

	pvzRes, err := s.pvzRepo.Create(ctx, domain.PVZ{
		ID:               createIn.ID,
		RegistrationDate: createIn.RegistrationDate,
		CityID:           city.ID,
	})
	if err != nil {
		if errors.Is(err, infra.ErrDuplicate) {
			return nil, domain.ErrDuplicatePvzID
		}
		return nil, fmt.Errorf("%s: failed to create pvz: %w", op, err)
	}

	return pvzRes, nil
}

func (s *PVZUseCase) List(ctx context.Context, pvzListParams *dto.PVZListParams) ([]*domain.PVZ, error) {
	const op = "pvz.List"

	pvzEnts, err := s.pvzRepo.ListPvzByAcceptanceDateAndCity(ctx, pvzListParams.Pagination, pvzListParams.Filter.StartDate, pvzListParams.Filter.EndDate)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get list pvz: %w", op, err)
	}

	if len(pvzEnts) == 0 {
		return []*domain.PVZ{}, nil
	}

	pvzIDs := make([]uuid.UUID, 0, len(pvzEnts))
	for _, pvzEnt := range pvzEnts {
		pvzIDs = append(pvzIDs, pvzEnt.ID)
	}

	receptionEnts, err := s.receptionRepo.ListByIDsWithStatus(ctx, pvzIDs)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get list receptions: %w", op, err)
	}

	receptionIDs := make([]uuid.UUID, 0, len(pvzEnts))
	mapPvzIDReceptions := make(map[uuid.UUID][]*domain.Reception, len(receptionEnts))
	for _, receptionEnt := range receptionEnts {
		receptionIDs = append(receptionIDs, receptionEnt.ID)
		mapPvzIDReceptions[receptionEnt.PvzID] = append(mapPvzIDReceptions[receptionEnt.PvzID], receptionEnt)
	}

	productEnts, err := s.productRepo.ListByReceptionIDsWithTypeName(ctx, receptionIDs)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get list products: %w", op, err)
	}

	mapReceptionIDProducts := make(map[uuid.UUID][]*domain.Product, len(productEnts))
	for _, productEnt := range productEnts {
		mapReceptionIDProducts[productEnt.ReceptionID] = append(mapReceptionIDProducts[productEnt.ReceptionID], productEnt)
	}

	outs := make([]*domain.PVZ, 0, len(pvzEnts))

	for _, pvz := range pvzEnts {
		pvzReceptions, ok := mapPvzIDReceptions[pvz.ID]
		if !ok {
			outs = append(outs, &domain.PVZ{
				ID:               pvz.ID,
				RegistrationDate: pvz.RegistrationDate,
				CityID:           pvz.CityID,
				Receptions:       nil,
			})
			continue
		}

		receptionsWithProducts := make([]*domain.Reception, 0, len(pvzReceptions))

		for _, reception := range pvzReceptions {
			productsWithTypeName := mapReceptionIDProducts[reception.ID]

			receptionsWithProducts = append(receptionsWithProducts, &domain.Reception{
				ID:       reception.ID,
				PvzID:    reception.PvzID,
				DateTime: reception.DateTime,
				StatusID: reception.StatusID,
				Products: productsWithTypeName,
			})
		}

		outs = append(outs, &domain.PVZ{
			ID:               pvz.ID,
			RegistrationDate: pvz.RegistrationDate,
			CityID:           pvz.CityID,
			Receptions:       receptionsWithProducts,
		})
	}

	return outs, nil
}
