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

func TestCityRepository_Create(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx postgres.DBTX) {
		cityRepo := postgres.NewCityRepository(tx)

		city1 := domain.City{
			ID:   uuid.New(),
			Name: "CityOne",
		}

		created1, err := cityRepo.Create(ctx, city1)
		require.NoError(t, err)
		require.NotNil(t, created1)
		assert.Equal(t, city1.ID, created1.ID)
		assert.Equal(t, city1.Name, created1.Name)

		city2 := domain.City{
			ID:   uuid.Nil,
			Name: "CityTwo",
		}

		created2, err := cityRepo.Create(ctx, city2)
		require.NoError(t, err)
		require.NotNil(t, created2)
		assert.NotEqual(t, uuid.Nil, created2.ID)
		assert.Equal(t, city2.Name, created2.Name)
	})
}

func TestCityRepository_Get(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx postgres.DBTX) {
		cityRepo := postgres.NewCityRepository(tx)

		// создаём город напрямую
		id := uuid.New()
		name := "TestCity"
		_, err := tx.Exec(ctx, `INSERT INTO cities (id, name) VALUES ($1, $2)`, id, name)
		require.NoError(t, err)

		tests := []struct {
			name     string
			filter   domain.City
			wantID   uuid.UUID
			wantName string
			wantErr  bool
		}{
			{"Get by ID", domain.City{ID: id}, id, name, false},
			{"Get by Name", domain.City{Name: name}, id, name, false},
			{"Not found", domain.City{ID: uuid.New()}, uuid.Nil, "", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := cityRepo.Get(ctx, tt.filter)
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

func TestCityRepository_CreateBatch(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx postgres.DBTX) {
		cityRepo := postgres.NewCityRepository(tx)

		tests := []struct {
			name    string
			cities  []domain.City
			wantErr bool
		}{
			{
				name: "Insert multiple cities",
				cities: []domain.City{
					{Name: "City A"},
					{Name: "City B"},
					{Name: "City C"},
				},
				wantErr: false,
			},
			{
				name:    "Insert empty slice",
				cities:  []domain.City{},
				wantErr: true, // пустой срез будет выдавать ошибку в текущем CreateBatch
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := cityRepo.CreateBatch(ctx, tt.cities)
				if tt.wantErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)

				// Проверяем, что города вставились (только для непустого среза)
				for _, city := range tt.cities {
					got, err := cityRepo.Get(ctx, domain.City{Name: city.Name})
					require.NoError(t, err)
					require.NotNil(t, got)
					assert.Equal(t, city.Name, got.Name)
					assert.NotEqual(t, uuid.Nil, got.ID)
				}
			})
		}
	})
}
