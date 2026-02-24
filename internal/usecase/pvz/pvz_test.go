package pvz

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/infra"
	"github.com/valeragav/avito-pvz-service/internal/usecase/dto"
	"github.com/valeragav/avito-pvz-service/internal/usecase/pvz/mocks"
	"github.com/valeragav/avito-pvz-service/pkg/listparams"
	"github.com/valeragav/avito-pvz-service/pkg/testutils"
	"go.uber.org/mock/gomock"
)

type pvzMocks struct {
	MockPvzRepo       *mocks.MockpvzRepo
	MockCityRepo      *mocks.MockcityRepo
	MockReceptionRepo *mocks.MockreceptionRepo
	MockProductRepo   *mocks.MockproductRepo
}

func newPvZMocks(t *testing.T) *pvzMocks {
	ctrl := gomock.NewController(t)

	return &pvzMocks{
		MockPvzRepo:       mocks.NewMockpvzRepo(ctrl),
		MockCityRepo:      mocks.NewMockcityRepo(ctrl),
		MockReceptionRepo: mocks.NewMockreceptionRepo(ctrl),
		MockProductRepo:   mocks.NewMockproductRepo(ctrl),
	}
}

func TestPVZUseCase_Create(t *testing.T) {
	t.Parallel()

	testutils.InitTestLogger()
	ctx := context.Background()

	type fields struct {
		name    string
		req     dto.PVZCreate
		mockFn  func(f fields, m *pvzMocks)
		wantErr error
	}

	testcases := []fields{
		{
			name: "ok",
			req: dto.PVZCreate{
				ID:               uuid.New(),
				RegistrationDate: time.Now(),
				CityName:         "Moscow",
			},
			mockFn: func(f fields, m *pvzMocks) {
				city := &domain.City{
					ID:   uuid.New(),
					Name: f.req.CityName,
				}

				m.MockCityRepo.EXPECT().
					Get(ctx, domain.City{Name: f.req.CityName}).
					Return(city, nil).
					Times(1)

				m.MockPvzRepo.EXPECT().
					Create(ctx, domain.PVZ{
						ID:               f.req.ID,
						RegistrationDate: f.req.RegistrationDate,
						CityID:           city.ID,
					}).
					Return(&domain.PVZ{
						ID:               f.req.ID,
						RegistrationDate: f.req.RegistrationDate,
						CityID:           city.ID,
					}, nil).
					Times(1)
			},
			wantErr: nil,
		},
		{
			name: "failed to get city",
			req: dto.PVZCreate{
				ID:               uuid.New(),
				RegistrationDate: time.Now(),
				CityName:         "Moscow",
			},
			mockFn: func(f fields, m *pvzMocks) {
				m.MockCityRepo.EXPECT().
					Get(ctx, domain.City{Name: f.req.CityName}).
					Return(nil, errors.New("db error")).
					Times(1)
			},
			wantErr: errors.New("pvz.Create: failed to get city: db error"),
		},
		{
			name: "failed to create pvz",
			req: dto.PVZCreate{
				ID:               uuid.New(),
				RegistrationDate: time.Now(),
				CityName:         "Moscow",
			},
			mockFn: func(f fields, m *pvzMocks) {
				city := &domain.City{
					ID:   uuid.New(),
					Name: f.req.CityName,
				}

				m.MockCityRepo.EXPECT().
					Get(ctx, domain.City{Name: f.req.CityName}).
					Return(city, nil).
					Times(1)

				m.MockPvzRepo.EXPECT().
					Create(ctx, gomock.Any()).
					Return(nil, errors.New("create error")).
					Times(1)
			},
			wantErr: errors.New("pvz.Create: failed to create pvz: create error"),
		},

		{
			name: "failed to create pvz due to duplicate id",
			req: dto.PVZCreate{
				ID:               uuid.New(),
				RegistrationDate: time.Now(),
				CityName:         "Moscow",
			},
			mockFn: func(f fields, m *pvzMocks) {
				city := &domain.City{
					ID:   uuid.New(),
					Name: f.req.CityName,
				}

				m.MockCityRepo.EXPECT().
					Get(ctx, domain.City{Name: f.req.CityName}).
					Return(city, nil).
					Times(1)

				m.MockPvzRepo.EXPECT().
					Create(ctx, gomock.Any()).
					Return(nil, infra.ErrDuplicate).
					Times(1)
			},
			wantErr: domain.ErrDuplicatePvzID,
		},
		{
			name: "not found city",
			req: dto.PVZCreate{
				ID:               uuid.New(),
				RegistrationDate: time.Now(),
				CityName:         "Moscow",
			},
			mockFn: func(f fields, m *pvzMocks) {
				m.MockCityRepo.EXPECT().
					Get(ctx, domain.City{Name: f.req.CityName}).
					Return(nil, infra.ErrNotFound).
					Times(1)
			},
			wantErr: domain.ErrCityNotFound,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pvzMocks := newPvZMocks(t)
			tt.mockFn(tt, pvzMocks)

			useCase := New(
				pvzMocks.MockPvzRepo,
				pvzMocks.MockCityRepo,
				pvzMocks.MockReceptionRepo,
				pvzMocks.MockProductRepo,
			)

			pvzRes, err := useCase.Create(ctx, tt.req)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr.Error())
				require.Nil(t, pvzRes)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, pvzRes)

			require.Equal(t, tt.req.ID, pvzRes.ID)
			require.Equal(t, tt.req.RegistrationDate, pvzRes.RegistrationDate)
		})
	}
}

func TestPVZUseCase_List(t *testing.T) {
	t.Parallel()

	testutils.InitTestLogger()
	ctx := context.Background()

	startDate := time.Now().Add(-24 * time.Hour)
	endDate := time.Now()

	params := &dto.PVZListParams{
		Pagination: &listparams.Pagination{
			Limit: 10,
			Page:  0,
		},
		Filter: &dto.PVZFilter{
			StartDate: &startDate,
			EndDate:   &endDate,
		},
	}

	type fields struct {
		name    string
		mockFn  func(m *pvzMocks)
		wantErr error
		checkFn func(t *testing.T, result []*domain.PVZ)
	}

	testcases := []fields{
		{
			name: "repo error on pvz list",
			mockFn: func(m *pvzMocks) {
				m.MockPvzRepo.EXPECT().
					ListPvzByAcceptanceDateAndCity(ctx, params.Pagination, &startDate, &endDate).
					Return(nil, errors.New("db error")).
					Times(1)
			},
			wantErr: errors.New("pvz.List: failed to get list pvz: db error"),
		},
		{
			name: "empty pvz list",
			mockFn: func(m *pvzMocks) {
				m.MockPvzRepo.EXPECT().
					ListPvzByAcceptanceDateAndCity(ctx, params.Pagination, &startDate, &endDate).
					Return([]*domain.PVZ{}, nil).
					Times(1)
			},
			checkFn: func(t *testing.T, result []*domain.PVZ) {
				require.NotNil(t, result)
				require.Empty(t, result)
			},
		},
		{
			name: "reception repo error",
			mockFn: func(m *pvzMocks) {
				pvzID := uuid.New()

				m.MockPvzRepo.EXPECT().
					ListPvzByAcceptanceDateAndCity(ctx, params.Pagination, &startDate, &endDate).
					Return([]*domain.PVZ{
						{ID: pvzID},
					}, nil).
					Times(1)

				m.MockReceptionRepo.EXPECT().
					ListByIDsWithStatus(ctx, []uuid.UUID{pvzID}).
					Return(nil, errors.New("reception error")).
					Times(1)
			},
			wantErr: errors.New("pvz.List: failed to get list receptions: reception error"),
		},
		{
			name: "product repo error",
			mockFn: func(m *pvzMocks) {
				pvzID := uuid.New()
				receptionID := uuid.New()

				m.MockPvzRepo.EXPECT().
					ListPvzByAcceptanceDateAndCity(ctx, params.Pagination, &startDate, &endDate).
					Return([]*domain.PVZ{
						{ID: pvzID},
					}, nil).
					Times(1)

				m.MockReceptionRepo.EXPECT().
					ListByIDsWithStatus(ctx, []uuid.UUID{pvzID}).
					Return([]*domain.Reception{
						{ID: receptionID, PvzID: pvzID},
					}, nil).
					Times(1)

				m.MockProductRepo.EXPECT().
					ListByReceptionIDsWithTypeName(ctx, []uuid.UUID{receptionID}).
					Return(nil, errors.New("product error")).
					Times(1)
			},
			wantErr: errors.New("pvz.List: failed to get list products: product error"),
		},
		{
			name: "full success with mapping",
			mockFn: func(m *pvzMocks) {
				pvzID := uuid.New()
				receptionID := uuid.New()
				productID := uuid.New()

				pvzEnt := &domain.PVZ{
					ID:               pvzID,
					RegistrationDate: time.Now(),
					CityID:           uuid.New(),
				}

				receptionEnt := &domain.Reception{
					ID:       receptionID,
					PvzID:    pvzID,
					DateTime: time.Now(),
					StatusID: uuid.New(),
				}

				productEnt := &domain.Product{
					ID:          productID,
					ReceptionID: receptionID,
				}

				m.MockPvzRepo.EXPECT().
					ListPvzByAcceptanceDateAndCity(ctx, params.Pagination, &startDate, &endDate).
					Return([]*domain.PVZ{pvzEnt}, nil).
					Times(1)

				m.MockReceptionRepo.EXPECT().
					ListByIDsWithStatus(ctx, []uuid.UUID{pvzID}).
					Return([]*domain.Reception{receptionEnt}, nil).
					Times(1)

				m.MockProductRepo.EXPECT().
					ListByReceptionIDsWithTypeName(ctx, []uuid.UUID{receptionID}).
					Return([]*domain.Product{productEnt}, nil).
					Times(1)
			},
			checkFn: func(t *testing.T, result []*domain.PVZ) {
				require.Len(t, result, 1)

				pvz := result[0]
				require.Len(t, pvz.Receptions, 1)

				reception := pvz.Receptions[0]
				require.Len(t, reception.Products, 1)

				require.Equal(t, reception.ID, reception.Products[0].ReceptionID)
			},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pvzMocks := newPvZMocks(t)
			tt.mockFn(pvzMocks)

			useCase := New(
				pvzMocks.MockPvzRepo,
				pvzMocks.MockCityRepo,
				pvzMocks.MockReceptionRepo,
				pvzMocks.MockProductRepo,
			)

			result, err := useCase.List(ctx, params)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr.Error())
				require.Nil(t, result)
				return
			}

			require.NoError(t, err)

			if tt.checkFn != nil {
				tt.checkFn(t, result)
			}
		})
	}
}
