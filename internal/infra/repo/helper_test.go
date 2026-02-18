package repo_test

import (
	"context"
	"fmt"
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/infra"
	"github.com/valeragav/avito-pvz-service/internal/infra/repo"
)

func TestCollectRowsAndOneRow(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx infra.DBTX) {
		type city struct {
			ID   uuid.UUID `db:"id"`
			Name string    `db:"name"`
		}

		createCity := city{
			ID:   uuid.New(),
			Name: "TestCity",
		}

		_, err := tx.Exec(ctx, `INSERT INTO cities (id, name) VALUES ($1, $2)`, createCity.ID, createCity.Name)
		require.NoError(t, err)

		sqb := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

		// Builder для CollectRows
		builder := sqb.Select("id", "name").From("cities").Where(sq.Eq{"id": createCity.ID})

		// ---------------- CollectRows ----------------
		results, err := repo.CollectRows(ctx, tx, builder, pgx.RowToStructByName[city])
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, createCity.ID, results[0].ID)
		assert.Equal(t, createCity.Name, results[0].Name)

		// Builder для CollectOneRow
		builderOne := sqb.Select("id", "name").From("cities").Where(sq.Eq{"id": createCity.ID})

		// ---------------- CollectOneRow ----------------
		result, err := repo.CollectOneRow(ctx, tx, builderOne, pgx.RowToStructByName[city])
		require.NoError(t, err)
		assert.Equal(t, createCity.ID, result.ID)
		assert.Equal(t, createCity.Name, result.Name)

		// ---------------- NotFound ----------------
		_, err = repo.CollectOneRow(ctx, tx, sqb.Select("id").From("cities").Where(sq.Eq{"id": uuid.New()}), pgx.RowToStructByName[city])
		assert.ErrorIs(t, err, infra.ErrNotFound)
	})
}

type badBuilder struct{}

func (b badBuilder) ToSql() (string, []any, error) {
	return "", nil, fmt.Errorf("builder error")
}

func TestCollectRows_BuildError(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx infra.DBTX) {
		mapper := func(row pgx.CollectableRow) (domain.City, error) {
			return domain.City{}, nil
		}
		_, err := repo.CollectRows(ctx, tx, badBuilder{}, mapper)
		assert.ErrorIs(t, err, infra.ErrBuildQuery)
	})
}

func TestCollectOneRow_DuplicateError(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx infra.DBTX) {
		id := uuid.New()
		_, err := tx.Exec(ctx, `INSERT INTO cities (id, name) VALUES ($1, $2)`, id, "City1")
		require.NoError(t, err)

		// пытаемся вставить ту же запись напрямую, чтобы вызвать ошибку уникальности
		_, err = tx.Exec(ctx, `INSERT INTO cities (id, name) VALUES ($1, $2)`, id, "City1")
		require.Error(t, err)                         // проверяем, что есть ошибка
		assert.True(t, repo.IsDuplicateKeyError(err)) // и что это действительно дубликат
	})
}
