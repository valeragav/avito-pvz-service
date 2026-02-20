package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/infra"
	"github.com/valeragav/avito-pvz-service/internal/infra/postgres"
)

func TestReceptionStatusRepository_Get(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx postgres.DBTX) {
		receptionStatusRepo := postgres.NewReceptionStatusRepository(tx)

		statuses := []domain.ReceptionStatus{
			{ID: uuid.New(), Name: "Open"},
			{ID: uuid.New(), Name: "Close"},
		}

		err := receptionStatusRepo.CreateBatch(ctx, statuses)
		require.NoError(t, err)

		got, err := receptionStatusRepo.Get(ctx, domain.ReceptionStatus{ID: statuses[0].ID})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, statuses[0].ID, got.ID)
		assert.Equal(t, statuses[0].Name, got.Name)

		got, err = receptionStatusRepo.Get(ctx, domain.ReceptionStatus{Name: statuses[1].Name})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, statuses[1].ID, got.ID)
		assert.Equal(t, statuses[1].Name, got.Name)

		_, err = receptionStatusRepo.Get(ctx, domain.ReceptionStatus{ID: uuid.New()})
		assert.ErrorIs(t, err, infra.ErrNotFound)

		_, err = receptionStatusRepo.Get(ctx, domain.ReceptionStatus{Name: "NonExistent"})
		assert.ErrorIs(t, err, infra.ErrNotFound)
	})
}

func TestReceptionStatusRepository_CreateBatch(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx postgres.DBTX) {
		receptionStatusRepo := postgres.NewReceptionStatusRepository(tx)

		statuses := []domain.ReceptionStatus{
			{ID: uuid.New(), Name: "Open"},
			{ID: uuid.Nil, Name: "Close"},
		}

		err := receptionStatusRepo.CreateBatch(ctx, statuses)
		require.NoError(t, err)

		gotOpen, err := receptionStatusRepo.Get(ctx, domain.ReceptionStatus{Name: "Open"})
		require.NoError(t, err)
		assert.Equal(t, "Open", string(gotOpen.Name))
		assert.Equal(t, statuses[0].ID, gotOpen.ID)

		gotClose, err := receptionStatusRepo.Get(ctx, domain.ReceptionStatus{Name: "Close"})
		require.NoError(t, err)
		assert.Equal(t, "Close", string(gotClose.Name))
		assert.NotEqual(t, uuid.Nil, gotClose.ID)

		duplicate := []domain.ReceptionStatus{
			{ID: uuid.New(), Name: "Open"},
		}
		err = receptionStatusRepo.CreateBatch(ctx, duplicate)
		require.NoError(t, err)

		// Проверяем, что количество записей не увеличилось
		// проверяем существующие по именам
		gotOpen2, err := receptionStatusRepo.Get(ctx, domain.ReceptionStatus{Name: "Open"})
		require.NoError(t, err)
		assert.Equal(t, gotOpen.ID, gotOpen2.ID)

		gotClose2, err := receptionStatusRepo.Get(ctx, domain.ReceptionStatus{Name: "Close"})
		require.NoError(t, err)
		assert.Equal(t, gotClose.ID, gotClose2.ID)
	})
}
