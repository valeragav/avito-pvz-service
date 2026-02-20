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
)

func TestReceptionRepository_Create(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx postgres.DBTX) {
		cityRepo := postgres.NewCityRepository(tx)
		pvzRepo := postgres.NewPVZRepository(tx)
		statusRepo := postgres.NewReceptionStatusRepository(tx)
		receptionRepo := postgres.NewReceptionRepository(tx)

		city, err := cityRepo.Create(ctx, domain.City{ID: uuid.New(), Name: "TestCity"})
		require.NoError(t, err)

		pvz, err := pvzRepo.Create(ctx, domain.PVZ{
			ID:               uuid.New(),
			RegistrationDate: time.Now(),
			CityID:           city.ID,
		})
		require.NoError(t, err)

		statusID := uuid.New()
		err = statusRepo.CreateBatch(ctx, []domain.ReceptionStatus{{
			ID:   statusID,
			Name: domain.ReceptionStatusClose,
		}})
		require.NoError(t, err)

		reception1 := domain.Reception{
			ID:       uuid.New(),
			PvzID:    pvz.ID,
			DateTime: time.Now(),
			StatusID: statusID,
		}

		created1, err := receptionRepo.Create(ctx, reception1)
		require.NoError(t, err)
		require.NotNil(t, created1)
		assert.Equal(t, reception1.ID, created1.ID)
		assert.Equal(t, reception1.PvzID, created1.PvzID)
		assert.WithinDuration(t, reception1.DateTime, created1.DateTime, time.Millisecond)
		assert.Equal(t, reception1.StatusID, created1.StatusID)

		reception2 := domain.Reception{
			ID:       uuid.Nil,
			PvzID:    pvz.ID,
			DateTime: time.Now(),
			StatusID: statusID,
		}

		created2, err := receptionRepo.Create(ctx, reception2)
		require.NoError(t, err)
		require.NotNil(t, created2)
		assert.NotEqual(t, uuid.Nil, created2.ID) // ID сгенерирован
		assert.Equal(t, reception2.PvzID, created2.PvzID)
		assert.WithinDuration(t, reception2.DateTime, created2.DateTime, time.Millisecond)
		assert.Equal(t, reception2.StatusID, created2.StatusID)
	})
}

func TestReceptionRepository_GetList(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx postgres.DBTX) {
		cityRepo := postgres.NewCityRepository(tx)
		pvzRepo := postgres.NewPVZRepository(tx)
		statusRepo := postgres.NewReceptionStatusRepository(tx)
		receptionRepo := postgres.NewReceptionRepository(tx)

		city, err := cityRepo.Create(ctx, domain.City{ID: uuid.New(), Name: "TestCity"})
		require.NoError(t, err)

		pvz1, err := pvzRepo.Create(ctx, domain.PVZ{ID: uuid.New(), RegistrationDate: time.Now(), CityID: city.ID})
		require.NoError(t, err)

		pvz2, err := pvzRepo.Create(ctx, domain.PVZ{ID: uuid.New(), RegistrationDate: time.Now(), CityID: city.ID})
		require.NoError(t, err)

		statusID := uuid.New()
		err = statusRepo.CreateBatch(ctx, []domain.ReceptionStatus{{
			ID:   statusID,
			Name: domain.ReceptionStatusClose,
		}})
		require.NoError(t, err)

		receptions := []domain.Reception{
			{ID: uuid.New(), PvzID: pvz1.ID, DateTime: time.Now(), StatusID: statusID},
			{ID: uuid.New(), PvzID: pvz1.ID, DateTime: time.Now().Add(10 * time.Minute), StatusID: statusID},
			{ID: uuid.New(), PvzID: pvz2.ID, DateTime: time.Now(), StatusID: statusID},
		}

		for i := range receptions {
			_, err = receptionRepo.Create(ctx, receptions[i])
			require.NoError(t, err)
		}

		allReceptions, err := receptionRepo.GetList(ctx, domain.Reception{})
		require.NoError(t, err)
		assert.Len(t, allReceptions, len(receptions))

		filteredByPvz, err := receptionRepo.GetList(ctx, domain.Reception{PvzID: pvz1.ID})
		require.NoError(t, err)
		assert.Len(t, filteredByPvz, 2)
		for _, r := range filteredByPvz {
			assert.Equal(t, pvz1.ID, r.PvzID)
		}

		filteredByStatus, err := receptionRepo.GetList(ctx, domain.Reception{StatusID: statusID})
		require.NoError(t, err)
		assert.Len(t, filteredByStatus, len(receptions))
		for _, r := range filteredByStatus {
			assert.Equal(t, statusID, r.StatusID)
		}

		filteredCombined, err := receptionRepo.GetList(ctx, domain.Reception{PvzID: pvz2.ID, StatusID: statusID})
		require.NoError(t, err)
		assert.Len(t, filteredCombined, 1)
		assert.Equal(t, pvz2.ID, filteredCombined[0].PvzID)
		assert.Equal(t, statusID, filteredCombined[0].StatusID)
	})
}

func TestReceptionRepository_Get(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx postgres.DBTX) {
		cityRepo := postgres.NewCityRepository(tx)
		pvzRepo := postgres.NewPVZRepository(tx)
		statusRepo := postgres.NewReceptionStatusRepository(tx)
		receptionRepo := postgres.NewReceptionRepository(tx)

		city, err := cityRepo.Create(ctx, domain.City{ID: uuid.New(), Name: "TestCity"})
		require.NoError(t, err)

		pvz, err := pvzRepo.Create(ctx, domain.PVZ{ID: uuid.New(), RegistrationDate: time.Now(), CityID: city.ID})
		require.NoError(t, err)

		statusID := uuid.New()
		err = statusRepo.CreateBatch(ctx, []domain.ReceptionStatus{{
			ID:   statusID,
			Name: domain.ReceptionStatusClose,
		}})
		require.NoError(t, err)

		reception := domain.Reception{
			ID:       uuid.New(),
			PvzID:    pvz.ID,
			DateTime: time.Now(),
			StatusID: statusID,
		}
		created, err := receptionRepo.Create(ctx, reception)
		require.NoError(t, err)

		got, err := receptionRepo.Get(ctx, domain.Reception{ID: created.ID})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, created.ID, got.ID)
		assert.Equal(t, created.PvzID, got.PvzID)
		assert.WithinDuration(t, created.DateTime, got.DateTime, time.Millisecond)
		assert.Equal(t, created.StatusID, got.StatusID)

		got, err = receptionRepo.Get(ctx, domain.Reception{PvzID: pvz.ID})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, created.PvzID, got.PvzID)

		got, err = receptionRepo.Get(ctx, domain.Reception{StatusID: statusID})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, statusID, got.StatusID)

		_, err = receptionRepo.Get(ctx, domain.Reception{ID: uuid.New()})
		assert.ErrorIs(t, err, infra.ErrNotFound)
	})
}

func TestReceptionRepository_ListByIDsWithStatus(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx postgres.DBTX) {
		cityRepo := postgres.NewCityRepository(tx)
		pvzRepo := postgres.NewPVZRepository(tx)
		statusRepo := postgres.NewReceptionStatusRepository(tx)
		receptionRepo := postgres.NewReceptionRepository(tx)

		city, err := cityRepo.Create(ctx, domain.City{ID: uuid.New(), Name: "City"})
		require.NoError(t, err)

		pvz1, err := pvzRepo.Create(ctx, domain.PVZ{ID: uuid.New(), RegistrationDate: time.Now(), CityID: city.ID})
		require.NoError(t, err)

		pvz2, err := pvzRepo.Create(ctx, domain.PVZ{ID: uuid.New(), RegistrationDate: time.Now(), CityID: city.ID})
		require.NoError(t, err)

		statusID := uuid.New()
		statusName := domain.ReceptionStatusClose
		err = statusRepo.CreateBatch(ctx, []domain.ReceptionStatus{
			{ID: statusID, Name: statusName},
		})
		require.NoError(t, err)

		pvzIDs := make([]uuid.UUID, 0, 2)
		pvzIDs = append(pvzIDs, pvz1.ID, pvz2.ID)

		for i := range 2 {
			_, err = receptionRepo.Create(ctx, domain.Reception{
				ID:       uuid.New(),
				PvzID:    pvz1.ID,
				DateTime: time.Now().Add(time.Duration(i) * time.Minute),
				StatusID: statusID,
			})
			require.NoError(t, err)

			_, err = receptionRepo.Create(ctx, domain.Reception{
				ID:       uuid.New(),
				PvzID:    pvz2.ID,
				DateTime: time.Now().Add(time.Duration(i) * time.Minute),
				StatusID: statusID,
			})
			require.NoError(t, err)
		}

		got, err := receptionRepo.ListByIDsWithStatus(ctx, pvzIDs)
		require.NoError(t, err)
		assert.Len(t, got, 4) // по 2 приёма на каждый PVZ

		for _, r := range got {
			assert.Contains(t, pvzIDs, r.PvzID)
			assert.Equal(t, statusID, r.StatusID)
			assert.Equal(t, statusName, r.ReceptionStatus.Name)
		}

		empty, err := receptionRepo.ListByIDsWithStatus(ctx, []uuid.UUID{})
		require.NoError(t, err)
		assert.Empty(t, empty)

		nonexistentID := uuid.New()
		empty, err = receptionRepo.ListByIDsWithStatus(ctx, []uuid.UUID{nonexistentID})
		require.NoError(t, err)
		assert.Empty(t, empty)
	})
}

func TestReceptionRepository_FindByStatus(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx postgres.DBTX) {
		cityRepo := postgres.NewCityRepository(tx)
		pvzRepo := postgres.NewPVZRepository(tx)
		statusRepo := postgres.NewReceptionStatusRepository(tx)
		receptionRepo := postgres.NewReceptionRepository(tx)

		city, err := cityRepo.Create(ctx, domain.City{ID: uuid.New(), Name: "City"})
		require.NoError(t, err)

		pvz, err := pvzRepo.Create(ctx, domain.PVZ{ID: uuid.New(), RegistrationDate: time.Now(), CityID: city.ID})
		require.NoError(t, err)

		statusCloseID := uuid.New()
		statusCloseName := domain.ReceptionStatusClose
		statusOpenID := uuid.New()
		statusOpenName := domain.ReceptionStatusInProgress

		err = statusRepo.CreateBatch(ctx, []domain.ReceptionStatus{
			{ID: statusCloseID, Name: statusCloseName},
			{ID: statusOpenID, Name: statusOpenName},
		})
		require.NoError(t, err)

		now := time.Now()
		_, err = receptionRepo.Create(ctx, domain.Reception{
			ID:       uuid.New(),
			PvzID:    pvz.ID,
			DateTime: now.Add(-10 * time.Minute),
			StatusID: statusCloseID,
		})
		require.NoError(t, err)

		reception2, err := receptionRepo.Create(ctx, domain.Reception{
			ID:       uuid.New(),
			PvzID:    pvz.ID,
			DateTime: now.Add(-5 * time.Minute),
			StatusID: statusCloseID,
		})
		require.NoError(t, err)

		reception3, err := receptionRepo.Create(ctx, domain.Reception{
			ID:       uuid.New(),
			PvzID:    pvz.ID,
			DateTime: now,
			StatusID: statusOpenID,
		})
		require.NoError(t, err)

		got, err := receptionRepo.FindByStatus(ctx, statusCloseName, domain.Reception{PvzID: pvz.ID})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, reception2.ID, got.ID)
		assert.Equal(t, statusCloseID, got.StatusID)
		assert.Equal(t, statusCloseName, got.ReceptionStatus.Name)

		got, err = receptionRepo.FindByStatus(ctx, statusOpenName, domain.Reception{PvzID: pvz.ID})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, reception3.ID, got.ID)
		assert.Equal(t, statusOpenID, got.StatusID)
		assert.Equal(t, statusOpenName, got.ReceptionStatus.Name)

		_, err = receptionRepo.FindByStatus(ctx, "nonexistent_status", domain.Reception{PvzID: pvz.ID})
		assert.ErrorIs(t, err, infra.ErrNotFound)
	})
}

func TestReceptionRepository_Update(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx postgres.DBTX) {
		cityRepo := postgres.NewCityRepository(tx)
		pvzRepo := postgres.NewPVZRepository(tx)
		statusRepo := postgres.NewReceptionStatusRepository(tx)
		receptionRepo := postgres.NewReceptionRepository(tx)

		city, err := cityRepo.Create(ctx, domain.City{ID: uuid.New(), Name: "City"})
		require.NoError(t, err)

		pvz, err := pvzRepo.Create(ctx, domain.PVZ{ID: uuid.New(), RegistrationDate: time.Now(), CityID: city.ID})
		require.NoError(t, err)

		statusOldID := uuid.New()
		statusNewID := uuid.New()
		err = statusRepo.CreateBatch(ctx, []domain.ReceptionStatus{
			{ID: statusOldID, Name: domain.ReceptionStatusClose},
			{ID: statusNewID, Name: domain.ReceptionStatusInProgress},
		})
		require.NoError(t, err)

		created, err := receptionRepo.Create(ctx, domain.Reception{
			ID:       uuid.New(),
			PvzID:    pvz.ID,
			DateTime: time.Now().Add(-1 * time.Hour),
			StatusID: statusOldID,
		})
		require.NoError(t, err)

		newDateTime := time.Now()
		updated, err := receptionRepo.Update(ctx, created.ID, domain.Reception{
			DateTime: newDateTime,
			StatusID: statusNewID,
		})
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, created.ID, updated.ID)
		assert.Equal(t, created.PvzID, updated.PvzID)
		assert.Equal(t, statusNewID, updated.StatusID)
		assert.WithinDuration(t, newDateTime, updated.DateTime, time.Millisecond)

		_, err = receptionRepo.Update(ctx, uuid.New(), domain.Reception{
			StatusID: statusNewID,
		})
		assert.ErrorIs(t, err, infra.ErrNotFound)
	})
}
