package product

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/dto"
	"github.com/valeragav/avito-pvz-service/internal/infra"
	"github.com/valeragav/avito-pvz-service/internal/usecase/product/mocks"
	"github.com/valeragav/avito-pvz-service/pkg/testutils"
	"go.uber.org/mock/gomock"
)

type productMocks struct {
	MockProductRepo     *mocks.MockproductRepo
	MockReceptionRepo   *mocks.MockreceptionRepo
	MockProductTypeRepo *mocks.MockproductTypeRepo
	MockPvzRepo         *mocks.MockpvzRepo
}

func newProductMocks(t *testing.T) *productMocks {
	ctrl := gomock.NewController(t)

	return &productMocks{
		MockProductRepo:     mocks.NewMockproductRepo(ctrl),
		MockReceptionRepo:   mocks.NewMockreceptionRepo(ctrl),
		MockProductTypeRepo: mocks.NewMockproductTypeRepo(ctrl),
		MockPvzRepo:         mocks.NewMockpvzRepo(ctrl),
	}
}

func TestProductUseCase_Create(t *testing.T) {
	t.Parallel()

	testutils.InitTestLogger()
	ctx := context.Background()

	type fields struct {
		name    string
		req     dto.ProductCreate
		mockFn  func(f fields, m *productMocks)
		wantErr error
	}

	testcases := []fields{
		{
			name: "ok",
			req: dto.ProductCreate{
				PvzID:    uuid.New(),
				TypeName: "Electronics",
			},
			mockFn: func(f fields, m *productMocks) {
				lastReception := &domain.Reception{ID: uuid.New()}
				productType := &domain.ProductType{ID: uuid.New()}

				m.MockReceptionRepo.EXPECT().
					FindByStatus(ctx, domain.ReceptionStatusInProgress, domain.Reception{PvzID: f.req.PvzID}).
					Return(lastReception, nil).
					Times(1)

				m.MockProductTypeRepo.EXPECT().
					Get(ctx, domain.ProductType{Name: f.req.TypeName}).
					Return(productType, nil).
					Times(1)

				m.MockProductRepo.EXPECT().
					Create(ctx, gomock.Any()).
					DoAndReturn(func(ctx context.Context, p domain.Product) (*domain.Product, error) {
						p.ID = uuid.New()
						return &p, nil
					}).
					Times(1)
			},
			wantErr: nil,
		},
		{
			name: "no reception in progress",
			req: dto.ProductCreate{
				PvzID:    uuid.New(),
				TypeName: "Electronics",
			},
			mockFn: func(f fields, m *productMocks) {
				m.MockReceptionRepo.EXPECT().
					FindByStatus(ctx, domain.ReceptionStatusInProgress, domain.Reception{PvzID: f.req.PvzID}).
					Return(nil, infra.ErrNotFound).
					Times(1)
			},
			wantErr: domain.ErrNoReceptionIsCurrentlyInProgress,
		},
		{
			name: "reception repo error",
			req: dto.ProductCreate{
				PvzID:    uuid.New(),
				TypeName: "Electronics",
			},
			mockFn: func(f fields, m *productMocks) {
				m.MockReceptionRepo.EXPECT().
					FindByStatus(ctx, domain.ReceptionStatusInProgress, domain.Reception{PvzID: f.req.PvzID}).
					Return(nil, errors.New("db error")).
					Times(1)
			},
			wantErr: errors.New("products.Create: failed to find in progress reception: db error"),
		},
		{
			name: "product type not found",
			req: dto.ProductCreate{
				PvzID:    uuid.New(),
				TypeName: "Electronics",
			},
			mockFn: func(f fields, m *productMocks) {
				lastReception := &domain.Reception{ID: uuid.New()}
				m.MockReceptionRepo.EXPECT().
					FindByStatus(ctx, domain.ReceptionStatusInProgress, domain.Reception{PvzID: f.req.PvzID}).
					Return(lastReception, nil).
					Times(1)

				m.MockProductTypeRepo.EXPECT().
					Get(ctx, domain.ProductType{Name: f.req.TypeName}).
					Return(nil, errors.New("not found")).
					Times(1)
			},
			wantErr: errors.New("products.Create: failed to find product type 'Electronics': not found"),
		},
		{
			name: "repo create error",
			req: dto.ProductCreate{
				PvzID:    uuid.New(),
				TypeName: "Electronics",
			},
			mockFn: func(f fields, m *productMocks) {
				lastReception := &domain.Reception{ID: uuid.New()}
				productType := &domain.ProductType{ID: uuid.New()}

				m.MockReceptionRepo.EXPECT().
					FindByStatus(ctx, domain.ReceptionStatusInProgress, domain.Reception{PvzID: f.req.PvzID}).
					Return(lastReception, nil).
					Times(1)

				m.MockProductTypeRepo.EXPECT().
					Get(ctx, domain.ProductType{Name: f.req.TypeName}).
					Return(productType, nil).
					Times(1)

				m.MockProductRepo.EXPECT().
					Create(ctx, gomock.Any()).
					Return(nil, errors.New("db error")).
					Times(1)
			},
			wantErr: errors.New("products.Create: failed to create product: db error"),
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			productMocks := newProductMocks(t)
			tt.mockFn(tt, productMocks)

			useCase := New(
				productMocks.MockProductRepo,
				productMocks.MockReceptionRepo,
				productMocks.MockProductTypeRepo,
				productMocks.MockPvzRepo,
			)

			product, err := useCase.Create(ctx, tt.req)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr.Error())
				require.Nil(t, product)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, product)

			if product.ID == uuid.Nil {
				require.Error(t, errors.New("id product empty"))
			}
		})
	}
}

func TestProductUseCase_DeleteLastProduct(t *testing.T) {
	t.Parallel()

	testutils.InitTestLogger()
	ctx := context.Background()

	type fields struct {
		name    string
		pvzID   uuid.UUID
		mockFn  func(f fields, m *productMocks)
		wantErr error
	}

	testcases := []fields{
		{
			name:  "ok",
			pvzID: uuid.New(),
			mockFn: func(f fields, m *productMocks) {
				pvz := &domain.PVZ{ID: f.pvzID}
				lastReception := &domain.Reception{ID: uuid.New()}
				lastProduct := &domain.Product{ID: uuid.New()}

				m.MockPvzRepo.EXPECT().
					Get(ctx, domain.PVZ{ID: f.pvzID}).
					Return(pvz, nil).
					Times(1)

				m.MockReceptionRepo.EXPECT().
					FindByStatus(ctx, domain.ReceptionStatusInProgress, domain.Reception{PvzID: f.pvzID}).
					Return(lastReception, nil).
					Times(1)

				m.MockProductRepo.EXPECT().
					GetLastProductInReception(ctx, lastReception.ID).
					Return(lastProduct, nil).
					Times(1)

				m.MockProductRepo.EXPECT().
					DeleteProduct(ctx, lastProduct.ID).
					Return(nil).
					Times(1)
			},
			wantErr: nil,
		},
		{
			name:  "pvz not found",
			pvzID: uuid.New(),
			mockFn: func(f fields, m *productMocks) {
				m.MockPvzRepo.EXPECT().
					Get(ctx, domain.PVZ{ID: f.pvzID}).
					Return(nil, infra.ErrNotFound).
					Times(1)
			},
			wantErr: domain.ErrPVZNotFound,
		},
		{
			name:  "reception not found",
			pvzID: uuid.New(),
			mockFn: func(f fields, m *productMocks) {
				m.MockPvzRepo.EXPECT().
					Get(ctx, domain.PVZ{ID: f.pvzID}).
					Return(&domain.PVZ{ID: f.pvzID}, nil).
					Times(1)

				m.MockReceptionRepo.EXPECT().
					FindByStatus(ctx, domain.ReceptionStatusInProgress, domain.Reception{PvzID: f.pvzID}).
					Return(nil, infra.ErrNotFound).
					Times(1)
			},
			wantErr: domain.ErrNoReceptionIsCurrentlyInProgress,
		},
		{
			name:  "last product not found",
			pvzID: uuid.New(),
			mockFn: func(f fields, m *productMocks) {
				lastReception := &domain.Reception{ID: uuid.New()}

				m.MockPvzRepo.EXPECT().
					Get(ctx, domain.PVZ{ID: f.pvzID}).
					Return(&domain.PVZ{ID: f.pvzID}, nil).
					Times(1)

				m.MockReceptionRepo.EXPECT().
					FindByStatus(ctx, domain.ReceptionStatusInProgress, domain.Reception{PvzID: f.pvzID}).
					Return(lastReception, nil).
					Times(1)

				m.MockProductRepo.EXPECT().
					GetLastProductInReception(ctx, lastReception.ID).
					Return(nil, errors.New("not found")).
					Times(1)
			},
			wantErr: errors.New("products.DeleteLastProduct: failed to get last product: not found"),
		},
		{
			name:  "delete product error",
			pvzID: uuid.New(),
			mockFn: func(f fields, m *productMocks) {
				lastReception := &domain.Reception{ID: uuid.New()}
				lastProduct := &domain.Product{ID: uuid.New()}

				m.MockPvzRepo.EXPECT().
					Get(ctx, domain.PVZ{ID: f.pvzID}).
					Return(&domain.PVZ{ID: f.pvzID}, nil).
					Times(1)

				m.MockReceptionRepo.EXPECT().
					FindByStatus(ctx, domain.ReceptionStatusInProgress, domain.Reception{PvzID: f.pvzID}).
					Return(lastReception, nil).
					Times(1)

				m.MockProductRepo.EXPECT().
					GetLastProductInReception(ctx, lastReception.ID).
					Return(lastProduct, nil).
					Times(1)

				m.MockProductRepo.EXPECT().
					DeleteProduct(ctx, lastProduct.ID).
					Return(errors.New("delete error")).
					Times(1)
			},
			wantErr: errors.New("products.DeleteLastProduct: failed to delete product: delete error"),
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			productMocks := newProductMocks(t)
			tt.mockFn(tt, productMocks)

			useCase := New(
				productMocks.MockProductRepo,
				productMocks.MockReceptionRepo,
				productMocks.MockProductTypeRepo,
				productMocks.MockPvzRepo,
			)

			product, err := useCase.DeleteLastProduct(ctx, tt.pvzID)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr.Error())
				require.Nil(t, product)
				return
			}

			require.NoError(t, err)
		})
	}
}
