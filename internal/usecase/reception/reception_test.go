package reception

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/dto"
	"github.com/valeragav/avito-pvz-service/internal/infra"
	"github.com/valeragav/avito-pvz-service/internal/usecase/reception/mocks"
	"github.com/valeragav/avito-pvz-service/pkg/testutils"
	"go.uber.org/mock/gomock"
)

type receptionMocks struct {
	MockReceptionRepo       *mocks.MockreceptionRepo
	MockReceptionStatusRepo *mocks.MockreceptionStatusRepo
	MockPvzRepo             *mocks.MockpvzRepo
}

func newReceptionMocks(t *testing.T) *receptionMocks {
	ctrl := gomock.NewController(t)

	return &receptionMocks{
		MockReceptionRepo:       mocks.NewMockreceptionRepo(ctrl),
		MockReceptionStatusRepo: mocks.NewMockreceptionStatusRepo(ctrl),
		MockPvzRepo:             mocks.NewMockpvzRepo(ctrl),
	}
}

func TestReceptionUseCase_Create(t *testing.T) {
	t.Parallel()

	testutils.InitTestLogger()
	ctx := context.Background()

	type fields struct {
		name    string
		req     dto.ReceptionCreate
		mockFn  func(f fields, m *receptionMocks)
		wantErr error
	}

	testcases := []fields{
		{
			name: "ok",
			req: dto.ReceptionCreate{
				PvzID: uuid.New(),
			},
			mockFn: func(f fields, m *receptionMocks) {
				statusID := uuid.New()

				m.MockReceptionRepo.EXPECT().
					FindByStatus(ctx, domain.ReceptionStatusInProgress, domain.Reception{
						PvzID: f.req.PvzID,
					}).
					Return(nil, infra.ErrNotFound).
					Times(1)

				m.MockReceptionStatusRepo.EXPECT().
					Get(ctx, domain.ReceptionStatus{
						Name: domain.ReceptionStatusInProgress,
					}).
					Return(&domain.ReceptionStatus{
						ID:   statusID,
						Name: domain.ReceptionStatusInProgress,
					}, nil).
					Times(1)

				m.MockReceptionRepo.EXPECT().
					Create(ctx, gomock.Any()).
					DoAndReturn(func(ctx context.Context, r domain.Reception) (*domain.Reception, error) {
						r.ID = uuid.New()
						return &r, nil
					}).
					Times(1)
			},
			wantErr: nil,
		},
		{
			name: "previous reception not found (business error)",
			req: dto.ReceptionCreate{
				PvzID: uuid.New(),
			},
			mockFn: func(f fields, m *receptionMocks) {
				m.MockReceptionRepo.EXPECT().
					FindByStatus(ctx, domain.ReceptionStatusInProgress, domain.Reception{
						PvzID: f.req.PvzID,
					}).
					Return(nil, nil).
					Times(1)
			},
			wantErr: domain.ErrNoReceptionIsCurrentlyInProgress,
		},
		{
			name: "failed to check last reception status",
			req: dto.ReceptionCreate{
				PvzID: uuid.New(),
			},
			mockFn: func(f fields, m *receptionMocks) {
				m.MockReceptionRepo.EXPECT().
					FindByStatus(ctx, domain.ReceptionStatusInProgress, domain.Reception{
						PvzID: f.req.PvzID,
					}).
					Return(nil, errors.New("db error")).
					Times(1)
			},
			wantErr: errors.New("receptions.Create: failed to check last reception status: db error"),
		},
		{
			name: "failed to get status",
			req: dto.ReceptionCreate{
				PvzID: uuid.New(),
			},
			mockFn: func(f fields, m *receptionMocks) {
				m.MockReceptionRepo.EXPECT().
					FindByStatus(ctx, domain.ReceptionStatusInProgress, domain.Reception{
						PvzID: f.req.PvzID,
					}).
					Return(nil, infra.ErrNotFound).
					Times(1)

				m.MockReceptionStatusRepo.EXPECT().
					Get(ctx, domain.ReceptionStatus{
						Name: domain.ReceptionStatusInProgress,
					}).
					Return(nil, errors.New("status error")).
					Times(1)
			},
			wantErr: errors.New("receptions.Create: failed to get status: status error"),
		},
		{
			name: "failed to create reception",
			req: dto.ReceptionCreate{
				PvzID: uuid.New(),
			},
			mockFn: func(f fields, m *receptionMocks) {
				statusID := uuid.New()

				m.MockReceptionRepo.EXPECT().
					FindByStatus(ctx, domain.ReceptionStatusInProgress, domain.Reception{
						PvzID: f.req.PvzID,
					}).
					Return(nil, infra.ErrNotFound).
					Times(1)

				m.MockReceptionStatusRepo.EXPECT().
					Get(ctx, domain.ReceptionStatus{
						Name: domain.ReceptionStatusInProgress,
					}).
					Return(&domain.ReceptionStatus{
						ID:   statusID,
						Name: domain.ReceptionStatusInProgress,
					}, nil).
					Times(1)

				m.MockReceptionRepo.EXPECT().
					Create(ctx, gomock.Any()).
					Return(nil, errors.New("create error")).
					Times(1)
			},
			wantErr: errors.New("receptions.Create: failed to create reception: create error"),
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			receptionMocks := newReceptionMocks(t)
			tt.mockFn(tt, receptionMocks)

			useCase := New(
				receptionMocks.MockReceptionRepo,
				receptionMocks.MockReceptionStatusRepo,
				receptionMocks.MockPvzRepo,
			)

			res, err := useCase.Create(ctx, tt.req)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr.Error())
				require.Nil(t, res)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, res)
			require.Equal(t, tt.req.PvzID, res.PvzID)
			require.NotZero(t, res.StatusID)
		})
	}
}

func TestReceptionUseCase_CloseLastReception(t *testing.T) {
	t.Parallel()

	testutils.InitTestLogger()
	ctx := context.Background()

	type fields struct {
		name    string
		pvzID   uuid.UUID
		mockFn  func(f fields, m *receptionMocks)
		wantErr error
	}

	testcases := []fields{
		{
			name:  "ok",
			pvzID: uuid.New(),
			mockFn: func(f fields, m *receptionMocks) {
				receptionID := uuid.New()
				statusID := uuid.New()

				m.MockPvzRepo.EXPECT().
					Get(ctx, domain.PVZ{ID: f.pvzID}).
					Return(&domain.PVZ{ID: f.pvzID}, nil).
					Times(1)

				m.MockReceptionRepo.EXPECT().
					FindByStatus(ctx, domain.ReceptionStatusInProgress, domain.Reception{
						PvzID: f.pvzID,
					}).
					Return(&domain.Reception{
						ID:    receptionID,
						PvzID: f.pvzID,
					}, nil).
					Times(1)

				m.MockReceptionStatusRepo.EXPECT().
					Get(ctx, domain.ReceptionStatus{
						Name: domain.ReceptionStatusClose,
					}).
					Return(&domain.ReceptionStatus{
						ID:   statusID,
						Name: domain.ReceptionStatusClose,
					}, nil).
					Times(1)

				m.MockReceptionRepo.EXPECT().
					Update(ctx, receptionID, domain.Reception{
						StatusID: statusID,
					}).
					Return(&domain.Reception{
						ID:       receptionID,
						PvzID:    f.pvzID,
						StatusID: statusID,
					}, nil).
					Times(1)
			},
			wantErr: nil,
		},
		{
			name:  "pvz not found",
			pvzID: uuid.New(),
			mockFn: func(f fields, m *receptionMocks) {
				m.MockPvzRepo.EXPECT().
					Get(ctx, domain.PVZ{ID: f.pvzID}).
					Return(nil, infra.ErrNotFound).
					Times(1)
			},
			wantErr: domain.ErrPVZNotFound,
		},
		{
			name:  "pvz repo error",
			pvzID: uuid.New(),
			mockFn: func(f fields, m *receptionMocks) {
				m.MockPvzRepo.EXPECT().
					Get(ctx, domain.PVZ{ID: f.pvzID}).
					Return(nil, errors.New("db error")).
					Times(1)
			},
			wantErr: errors.New("receptions.CloseLastReception: failed to find pvz: db error"),
		},
		{
			name:  "no reception in progress",
			pvzID: uuid.New(),
			mockFn: func(f fields, m *receptionMocks) {
				m.MockPvzRepo.EXPECT().
					Get(ctx, domain.PVZ{ID: f.pvzID}).
					Return(&domain.PVZ{ID: f.pvzID}, nil).
					Times(1)

				m.MockReceptionRepo.EXPECT().
					FindByStatus(ctx, domain.ReceptionStatusInProgress, domain.Reception{
						PvzID: f.pvzID,
					}).
					Return(nil, infra.ErrNotFound).
					Times(1)
			},
			wantErr: domain.ErrReceptionNotFound,
		},
		{
			name:  "reception repo error",
			pvzID: uuid.New(),
			mockFn: func(f fields, m *receptionMocks) {
				m.MockPvzRepo.EXPECT().
					Get(ctx, domain.PVZ{ID: f.pvzID}).
					Return(&domain.PVZ{ID: f.pvzID}, nil).
					Times(1)

				m.MockReceptionRepo.EXPECT().
					FindByStatus(ctx, domain.ReceptionStatusInProgress, domain.Reception{
						PvzID: f.pvzID,
					}).
					Return(nil, errors.New("reception error")).
					Times(1)
			},
			wantErr: errors.New("receptions.CloseLastReception: failed to find pvz: reception error"),
		},
		{
			name:  "failed to get close status",
			pvzID: uuid.New(),
			mockFn: func(f fields, m *receptionMocks) {
				receptionID := uuid.New()

				m.MockPvzRepo.EXPECT().
					Get(ctx, domain.PVZ{ID: f.pvzID}).
					Return(&domain.PVZ{ID: f.pvzID}, nil).
					Times(1)

				m.MockReceptionRepo.EXPECT().
					FindByStatus(ctx, domain.ReceptionStatusInProgress, domain.Reception{
						PvzID: f.pvzID,
					}).
					Return(&domain.Reception{ID: receptionID}, nil).
					Times(1)

				m.MockReceptionStatusRepo.EXPECT().
					Get(ctx, domain.ReceptionStatus{
						Name: domain.ReceptionStatusClose,
					}).
					Return(nil, errors.New("status error")).
					Times(1)
			},
			wantErr: errors.New("receptions.CloseLastReception: failed to get status: status error"),
		},
		{
			name:  "failed to update reception",
			pvzID: uuid.New(),
			mockFn: func(f fields, m *receptionMocks) {
				receptionID := uuid.New()
				statusID := uuid.New()

				m.MockPvzRepo.EXPECT().
					Get(ctx, domain.PVZ{ID: f.pvzID}).
					Return(&domain.PVZ{ID: f.pvzID}, nil).
					Times(1)

				m.MockReceptionRepo.EXPECT().
					FindByStatus(ctx, domain.ReceptionStatusInProgress, domain.Reception{
						PvzID: f.pvzID,
					}).
					Return(&domain.Reception{ID: receptionID}, nil).
					Times(1)

				m.MockReceptionStatusRepo.EXPECT().
					Get(ctx, domain.ReceptionStatus{
						Name: domain.ReceptionStatusClose,
					}).
					Return(&domain.ReceptionStatus{ID: statusID}, nil).
					Times(1)

				m.MockReceptionRepo.EXPECT().
					Update(ctx, receptionID, domain.Reception{
						StatusID: statusID,
					}).
					Return(nil, errors.New("update error")).
					Times(1)
			},
			wantErr: errors.New("failed to close reception: update error"),
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			receptionMocks := newReceptionMocks(t)
			tt.mockFn(tt, receptionMocks)

			useCase := New(
				receptionMocks.MockReceptionRepo,
				receptionMocks.MockReceptionStatusRepo,
				receptionMocks.MockPvzRepo,
			)

			res, err := useCase.CloseLastReception(ctx, tt.pvzID)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr.Error())
				require.Nil(t, res)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, res)
			require.Equal(t, tt.pvzID, res.PvzID)
		})
	}
}
