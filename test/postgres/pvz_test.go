package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/infra"
	"github.com/valeragav/avito-pvz-service/internal/infra/postgres"
	"github.com/valeragav/avito-pvz-service/pkg/listparams"
)

func TestPVZRepository_Create(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx postgres.DBTX) {
		cityRepo := postgres.NewCityRepository(tx)

		city, err := cityRepo.Create(ctx, domain.City{
			ID:   uuid.New(),
			Name: "TestCity",
		})
		require.NoError(t, err)
		require.NotNil(t, city)

		pvzRepo := postgres.NewPVZRepository(tx)

		now := time.Now()
		pvz := domain.PVZ{
			ID:               uuid.New(),
			RegistrationDate: now,
			CityID:           city.ID,
		}

		created, err := pvzRepo.Create(ctx, pvz)
		require.NoError(t, err)
		require.NotNil(t, created)

		assert.Equal(t, pvz.ID, created.ID)

		assert.WithinDuration(t, pvz.RegistrationDate, created.RegistrationDate, time.Millisecond)
		assert.Equal(t, pvz.CityID, created.CityID)
	})
}

func TestPVZRepository_Get(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx postgres.DBTX) {
		cityRepo := postgres.NewCityRepository(tx)

		city, err := cityRepo.Create(ctx, domain.City{
			ID:   uuid.New(),
			Name: "TestCity",
		})
		require.NoError(t, err)

		pvzRepo := postgres.NewPVZRepository(tx)

		pvz := domain.PVZ{
			ID:               uuid.New(),
			RegistrationDate: time.Now(),
			CityID:           city.ID,
		}

		created, err := pvzRepo.Create(ctx, pvz)
		require.NoError(t, err)
		require.NotNil(t, created)

		got, err := pvzRepo.Get(ctx, domain.PVZ{ID: pvz.ID})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, pvz.ID, got.ID)
		assert.WithinDuration(t, pvz.RegistrationDate, got.RegistrationDate, time.Millisecond)
		assert.Equal(t, pvz.CityID, got.CityID)

		_, err = pvzRepo.Get(ctx, domain.PVZ{ID: uuid.New()})
		assert.ErrorIs(t, err, infra.ErrNotFound)
	})
}

func TestPVZRepository_GetList(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx postgres.DBTX) {
		cityRepo := postgres.NewCityRepository(tx)

		city, err := cityRepo.Create(ctx, domain.City{
			ID:   uuid.New(),
			Name: "TestCity",
		})
		require.NoError(t, err)

		pvzRepo := postgres.NewPVZRepository(tx)

		pvzList := []domain.PVZ{
			{ID: uuid.New(), RegistrationDate: time.Now(), CityID: city.ID},
			{ID: uuid.New(), RegistrationDate: time.Now().Add(1 * time.Hour), CityID: city.ID},
		}

		for i := range pvzList {
			var created *domain.PVZ
			created, err = pvzRepo.Create(ctx, pvzList[i])
			require.NoError(t, err)
			pvzList[i].ID = created.ID
		}

		gotList, err := pvzRepo.GetList(ctx, nil)
		require.NoError(t, err)
		require.Len(t, gotList, len(pvzList))

		gotIDs := make(map[uuid.UUID]struct{})
		for _, pvz := range gotList {
			gotIDs[pvz.ID] = struct{}{}
		}

		for _, pvz := range pvzList {
			_, exists := gotIDs[pvz.ID]
			require.True(t, exists, "PVZ с ID %s должен быть в списке", pvz.ID)
		}

		// откатываем транзакцию, чтобы удалить все записи
		// или создаём новую транзакцию без вставок
		WithTx(t, func(ctx context.Context, tx postgres.DBTX) {
			emptyRepo := postgres.NewPVZRepository(tx)
			list, err := emptyRepo.GetList(ctx, nil)
			require.NoError(t, err)
			require.Empty(t, list)
		})
	})
}

func TestPVZRepository_ListPvzByAcceptanceDateAndCity(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx postgres.DBTX) {
		cityRepo := postgres.NewCityRepository(tx)
		receptionRepo := postgres.NewReceptionRepository(tx)
		pvzRepo := postgres.NewPVZRepository(tx)

		city1, err := cityRepo.Create(ctx, domain.City{ID: uuid.New(), Name: "City1"})
		require.NoError(t, err)

		city2, err := cityRepo.Create(ctx, domain.City{ID: uuid.New(), Name: "City2"})
		require.NoError(t, err)

		now := time.Now()
		pvz1, err := pvzRepo.Create(ctx, domain.PVZ{ID: uuid.New(), RegistrationDate: now.Add(-2 * time.Hour), CityID: city1.ID})
		require.NoError(t, err)

		pvz2, err := pvzRepo.Create(ctx, domain.PVZ{ID: uuid.New(), RegistrationDate: now.Add(-1 * time.Hour), CityID: city2.ID})
		require.NoError(t, err)

		pvz3, err := pvzRepo.Create(ctx, domain.PVZ{ID: uuid.New(), RegistrationDate: now.Add(-5 * time.Hour), CityID: city2.ID})
		require.NoError(t, err)

		receptionStatusID := uuid.New()
		receptionStatusRepo := postgres.NewReceptionStatusRepository(tx)
		err = receptionStatusRepo.CreateBatch(ctx, []domain.ReceptionStatus{{
			ID:   receptionStatusID,
			Name: domain.ReceptionStatusClose,
		}})
		require.NoError(t, err)

		_, err = receptionRepo.Create(ctx, domain.Reception{
			ID:       uuid.New(),
			PvzID:    pvz1.ID,
			DateTime: now.Add(-90 * time.Minute),
			StatusID: receptionStatusID,
		})
		require.NoError(t, err)

		_, err = receptionRepo.Create(ctx, domain.Reception{
			ID:       uuid.New(),
			PvzID:    pvz2.ID,
			DateTime: now.Add(-30 * time.Minute),
			StatusID: receptionStatusID,
		})
		require.NoError(t, err)

		_, err = receptionRepo.Create(ctx, domain.Reception{
			ID:       uuid.New(),
			PvzID:    pvz3.ID,
			DateTime: now.Add(-5 * time.Hour),
			StatusID: receptionStatusID,
		})
		require.NoError(t, err)

		pagination := &listparams.Pagination{Limit: 10, Page: 1}

		start := now.Add(-2 * time.Hour)
		end := now.Add(-1 * time.Minute)

		result, err := pvzRepo.ListPvzByAcceptanceDateAndCity(ctx, pagination, &start, &end)
		require.NoError(t, err)

		assert.Len(t, result, 2)

		ids := map[uuid.UUID]struct{}{}
		for _, pvz := range result {
			ids[pvz.ID] = struct{}{}
		}
		assert.Contains(t, ids, pvz1.ID, "PVZ1 должен быть в списке")
		assert.Contains(t, ids, pvz2.ID, "PVZ2 должен быть в списке")

		start = now.Add(-10 * time.Minute)
		end = now
		emptyResult, err := pvzRepo.ListPvzByAcceptanceDateAndCity(ctx, pagination, &start, &end)
		require.NoError(t, err)
		assert.Empty(t, emptyResult)
	})
}

func TestPVZRepository_ListPvzByAcceptanceDateAndCity_Pagination(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx postgres.DBTX) {
		cityRepo := postgres.NewCityRepository(tx)
		receptionRepo := postgres.NewReceptionRepository(tx)
		pvzRepo := postgres.NewPVZRepository(tx)

		city, err := cityRepo.Create(ctx, domain.City{ID: uuid.New(), Name: "City"})
		require.NoError(t, err)

		receptionStatusID := uuid.New()
		receptionStatusRepo := postgres.NewReceptionStatusRepository(tx)
		err = receptionStatusRepo.CreateBatch(ctx, []domain.ReceptionStatus{{
			ID:   receptionStatusID,
			Name: domain.ReceptionStatusClose,
		}})
		require.NoError(t, err)

		now := time.Now()
		pvzList := make([]*domain.PVZ, 5)
		for i := range 5 {
			var pvz *domain.PVZ
			pvz, err = pvzRepo.Create(ctx, domain.PVZ{
				ID:               uuid.New(),
				RegistrationDate: now.Add(time.Duration(-i) * time.Hour),
				CityID:           city.ID,
			})
			require.NoError(t, err)
			pvzList[i] = pvz

			_, err = receptionRepo.Create(ctx, domain.Reception{
				ID:       uuid.New(),
				PvzID:    pvz.ID,
				DateTime: now.Add(time.Duration(-i) * time.Hour),
				StatusID: receptionStatusID,
			})
			require.NoError(t, err)
		}

		pagination := &listparams.Pagination{Limit: 2, Page: 1}
		start := now.Add(-10 * time.Hour)
		end := now
		page1, err := pvzRepo.ListPvzByAcceptanceDateAndCity(ctx, pagination, &start, &end)
		require.NoError(t, err)
		assert.Len(t, page1, 2)

		// проверяем сортировку DESC
		assert.True(t, page1[0].RegistrationDate.After(page1[1].RegistrationDate) || page1[0].RegistrationDate.Equal(page1[1].RegistrationDate))

		pagination = &listparams.Pagination{Limit: 2, Page: 2}
		page2, err := pvzRepo.ListPvzByAcceptanceDateAndCity(ctx, pagination, &start, &end)
		require.NoError(t, err)
		assert.Len(t, page2, 2)
		assert.True(t, page2[0].RegistrationDate.After(page2[1].RegistrationDate) || page2[0].RegistrationDate.Equal(page2[1].RegistrationDate))

		pagination = &listparams.Pagination{Limit: 2, Page: 3}
		page3, err := pvzRepo.ListPvzByAcceptanceDateAndCity(ctx, pagination, &start, &end)
		require.NoError(t, err)
		assert.Len(t, page3, 1) // последняя страница содержит 1 элемент
	})
}
