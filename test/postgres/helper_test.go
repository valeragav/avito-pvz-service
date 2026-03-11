package postgres_test

import (
	"context"
	"errors"
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/infra"
	"github.com/valeragav/avito-pvz-service/internal/infra/postgres"
)

func TestCollectRowsAndOneRow(t *testing.T) {
	WithTx(t, func(ctx context.Context, qeProvider postgres.QueryEngineProvider) {
		type city struct {
			ID   uuid.UUID `db:"id"`
			Name string    `db:"name"`
		}

		createCity := city{
			ID:   uuid.New(),
			Name: "TestCity",
		}

		queryEngine := qeProvider.GetQueryEngine(ctx)

		_, err := queryEngine.Exec(ctx, `INSERT INTO cities (id, name) VALUES ($1, $2)`, createCity.ID, createCity.Name)
		require.NoError(t, err)

		sqb := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

		builder := sqb.Select("id", "name").From("cities").Where(sq.Eq{"id": createCity.ID})

		results, err := postgres.CollectRows(ctx, queryEngine, builder, pgx.RowToStructByName[city])
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, createCity.ID, results[0].ID)
		assert.Equal(t, createCity.Name, results[0].Name)

		builderOne := sqb.Select("id", "name").From("cities").Where(sq.Eq{"id": createCity.ID})

		result, err := postgres.CollectOneRow(ctx, queryEngine, builderOne, pgx.RowToStructByName[city])
		require.NoError(t, err)
		assert.Equal(t, createCity.ID, result.ID)
		assert.Equal(t, createCity.Name, result.Name)

		_, err = postgres.CollectOneRow(ctx, queryEngine, sqb.Select("id").From("cities").Where(sq.Eq{"id": uuid.New()}), pgx.RowToStructByName[city])
		assert.ErrorIs(t, err, infra.ErrNotFound)
	})
}

type badBuilder struct{}

func (b badBuilder) ToSql() (_ string, _ []any, err error) {
	err = errors.New("builder error")
	return
}

func TestCollectRows_BuildError(t *testing.T) {
	WithTx(t, func(ctx context.Context, qeProvider postgres.QueryEngineProvider) {
		mapper := func(row pgx.CollectableRow) (domain.City, error) {
			return domain.City{}, nil
		}
		_, err := postgres.CollectRows(ctx, qeProvider.GetQueryEngine(ctx), badBuilder{}, mapper)
		assert.ErrorIs(t, err, postgres.ErrBuildQuery)
	})
}

func TestCollectOneRow_DuplicateError(t *testing.T) {
	WithTx(t, func(ctx context.Context, qeProvider postgres.QueryEngineProvider) {
		id := uuid.New()
		_, err := qeProvider.GetQueryEngine(ctx).Exec(ctx, `INSERT INTO cities (id, name) VALUES ($1, $2)`, id, "City1")
		require.NoError(t, err)

		_, err = qeProvider.GetQueryEngine(ctx).Exec(ctx, `INSERT INTO cities (id, name) VALUES ($1, $2)`, id, "City1")
		require.Error(t, err)
		assert.True(t, postgres.IsDuplicateKeyError(err))
	})
}
