package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/infra/postgres"
)

func TestProductTypeRepository_Get(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx postgres.DBTX) {
		repo := postgres.NewProductTypeRepository(tx)

		id := uuid.New()
		name := "Test"
		_, err := tx.Exec(ctx, `INSERT INTO product_types (id, name) VALUES ($1, $2)`, id, name)
		require.NoError(t, err)

		tests := []struct {
			name     string
			filter   domain.ProductType
			wantID   uuid.UUID
			wantName string
			wantErr  bool
		}{
			{"Get by ID", domain.ProductType{ID: id}, id, name, false},
			{"Get by Name", domain.ProductType{Name: name}, id, name, false},
			{"Not found", domain.ProductType{ID: uuid.New()}, uuid.Nil, "", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := repo.Get(ctx, tt.filter)
				if tt.wantErr {
					require.Error(t, err)
					require.Nil(t, got)
				} else {
					require.NoError(t, err)
					require.NotNil(t, got)
					assert.Equal(t, tt.wantID, got.ID)
					assert.Equal(t, tt.wantName, got.Name)
				}
			})
		}
	})
}

func TestProductTypeRepository_CreateBatch(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx postgres.DBTX) {
		repo := postgres.NewProductTypeRepository(tx)

		tests := []struct {
			name         string
			productTypes []domain.ProductType
			wantErr      bool
		}{
			{
				name: "Insert multiple product types",
				productTypes: []domain.ProductType{
					{Name: "Electronics_test"},
					{Name: "Furniture_test"},
					{Name: "Clothing_test"},
				},
				wantErr: false,
			},
			{
				name:         "Insert empty slice",
				productTypes: []domain.ProductType{},
				wantErr:      true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := repo.CreateBatch(ctx, tt.productTypes)
				if tt.wantErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)

				for _, pt := range tt.productTypes {
					got, err := repo.Get(ctx, domain.ProductType{Name: pt.Name})
					require.NoError(t, err)
					require.NotNil(t, got)
					assert.Equal(t, pt.Name, got.Name)
					assert.NotEqual(t, uuid.Nil, got.ID)
				}
			})
		}
	})
}
