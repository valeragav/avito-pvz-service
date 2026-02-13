package pvz

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/valeragav/avito-pvz-service/internal/infrastructure/storage/cities"
	"github.com/valeragav/avito-pvz-service/internal/infrastructure/storage/products"
	"github.com/valeragav/avito-pvz-service/internal/infrastructure/storage/pvz"
	"github.com/valeragav/avito-pvz-service/internal/infrastructure/storage/receptions"
)

type PvzService struct {
	pvzRepo        *pvz.Repository
	citiesRepo     *cities.Repository
	receptionsRepo *receptions.Repository
	productsRepo   *products.Repository
}

func New(pvzRepo *pvz.Repository, citiesRepo *cities.Repository, receptionsRepo *receptions.Repository, productsRepo *products.Repository) *PvzService {
	return &PvzService{
		pvzRepo,
		citiesRepo,
		receptionsRepo,
		productsRepo,
	}
}

func (s *PvzService) Create(ctx context.Context, createIn CreateIn) (*CreateOut, error) {
	const op = "pvz.Create"

	city, err := s.citiesRepo.Get(ctx, cities.Cities{
		Name: createIn.CityName,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get city: %w", op, err)
	}

	pvzRes, err := s.pvzRepo.Create(ctx, pvz.Pvz{
		ID:               createIn.ID,
		RegistrationDate: createIn.RegistrationDate,
		CityID:           city.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create pvz: %w", op, err)
	}

	return &CreateOut{
		ID:               pvzRes.ID,
		CityName:         city.Name,
		RegistrationDate: createIn.RegistrationDate,
	}, nil
}

func (s *PvzService) List(ctx context.Context, pvzListParams *PvzListParams) (*PvzListResponse, error) {
	const op = "pvz.List"

	pvzEnts, err := s.pvzRepo.ListPvzByAcceptanceDateAndCity(ctx, pvzListParams.Pagination, pvzListParams.Filter.StartDate, pvzListParams.Filter.EndDate)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get list pvz: %w", op, err)
	}

	if len(pvzEnts) == 0 {
		return &PvzListResponse{}, nil
	}

	pvzIDs := make([]uuid.UUID, 0, len(pvzEnts))
	for _, pvzEnt := range pvzEnts {
		pvzIDs = append(pvzIDs, pvzEnt.ID)
	}

	receptionEnts, err := s.receptionsRepo.ListByIDsWithStatus(ctx, pvzIDs)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get list receptions: %w", op, err)
	}

	receptionIDs := make([]uuid.UUID, 0, len(pvzEnts))
	mapPvzIDReceptions := make(map[uuid.UUID][]receptions.ReceptionsWithStatus, len(receptionEnts))
	for _, receptionEnt := range receptionEnts {
		receptionIDs = append(receptionIDs, receptionEnt.ID)
		mapPvzIDReceptions[receptionEnt.PvzID] = append(mapPvzIDReceptions[receptionEnt.PvzID], receptionEnt)
	}

	productEnts, err := s.productsRepo.ListByReceptionIDsWithTypeName(ctx, receptionIDs)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get list products: %w", op, err)
	}

	mapReceptionIDProducts := make(map[uuid.UUID][]products.ProductsWithTypeName, len(productEnts))
	for _, productEnt := range productEnts {
		mapReceptionIDProducts[productEnt.ReceptionID] = append(mapReceptionIDProducts[productEnt.ReceptionID], productEnt)
	}

	outs := make([]Out, 0, len(pvzEnts))

	for _, pvz := range pvzEnts {
		pvzReceptions, ok := mapPvzIDReceptions[pvz.ID]
		if !ok {
			outs = append(outs, Out{
				Pvz:        pvz,
				Receptions: []ReceptionsWithProduct{},
			})
			continue
		}

		receptionsWithProducts := make([]ReceptionsWithProduct, 0, len(pvzReceptions))

		for _, reception := range pvzReceptions {
			productsWithTypeName := mapReceptionIDProducts[reception.ID]

			receptionsWithProducts = append(receptionsWithProducts, ReceptionsWithProduct{
				Reception: reception,
				Products:  productsWithTypeName,
			})
		}

		outs = append(outs, Out{
			Pvz:        pvz,
			Receptions: receptionsWithProducts,
		})
	}
	return &PvzListResponse{outs}, nil
}
