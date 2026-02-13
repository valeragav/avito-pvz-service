package pvz

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/VaLeraGav/avito-pvz-service/internal/infrastructure/storage"
	"github.com/VaLeraGav/avito-pvz-service/internal/infrastructure/storage/receptions"
	"github.com/VaLeraGav/avito-pvz-service/pkg/listparams"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db  *pgxpool.Pool
	sqb sq.StatementBuilderType
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{
		db:  db,
		sqb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r Repository) Create(ctx context.Context, pvz Pvz) (*Pvz, error) {
	if pvz.ID == uuid.Nil {
		pvz.ID = uuid.New()
	}

	queryBuilder := r.sqb.
		Insert(pvz.TableName()).
		Columns(Cols.ID, Cols.CityID, Cols.RegistrationDate).
		Values(pvz.ID, pvz.CityID, pvz.RegistrationDate).
		Suffix("RETURNING *")

	sql, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrBuildQuery, err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrExecuteQuery, err)
	}
	defer rows.Close()

	pvzCreate, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Pvz])
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrScanResult, err)
	}

	return &pvzCreate, nil
}

func (r *Repository) Get(ctx context.Context, filter Pvz) (*Pvz, error) {
	where := sq.Eq{}
	if filter.ID != uuid.Nil {
		where[Cols.ID] = filter.ID
	}
	if !filter.RegistrationDate.IsZero() {
		where[Cols.RegistrationDate] = filter.RegistrationDate
	}
	if filter.CityID != uuid.Nil {
		where[Cols.CityID] = filter.CityID
	}

	selectBuilder := r.sqb.
		Select(filter.AllCols()...).
		From(filter.TableName()).
		Where(where)

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrBuildQuery, err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrExecuteQuery, err)
	}
	defer rows.Close()

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Pvz])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, fmt.Errorf("%w: %w", storage.ErrScanResult, err)
	}

	return &result, nil
}

// TODO: передавать dto
func (r *Repository) ListPvzByAcceptanceDate(ctx context.Context, pagination listparams.Pagination, startDate, endDate *time.Time) ([]Pvz, error) {
	pvzAlias := "p"
	receptionsAlias := "r"

	pvzTable := Pvz{}.TableName()
	receptionsTable := receptions.Receptions{}.TableName()

	queryBuilder := r.sqb.
		Select(Pvz{}.AliasedCols(pvzAlias)...).
		From(pvzTable + " " + pvzAlias)

	if startDate != nil || endDate != nil {
		queryBuilder = queryBuilder.Join(
			fmt.Sprintf("%s %s ON %s.id = %s.pvz_id", receptionsTable, receptionsAlias, pvzAlias, receptionsAlias),
		)

		dataTimeCol := receptionsAlias + "." + receptions.Cols.DateTime
		if startDate != nil {
			queryBuilder = queryBuilder.Where(sq.GtOrEq{dataTimeCol: startDate})
		}

		if endDate != nil {
			queryBuilder = queryBuilder.Where(sq.LtOrEq{dataTimeCol: endDate})
		}

		queryBuilder = queryBuilder.GroupBy(pvzAlias + ".id")
	}

	offset := pagination.Offset()
	queryBuilder = queryBuilder.
		OrderBy(pvzAlias + ".registration_date DESC").
		Offset(uint64(offset)).
		Limit(uint64(pagination.Limit))

	sql, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrBuildQuery, err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrExecuteQuery, err)
	}
	defer rows.Close()

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[Pvz])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", storage.ErrNotFound, err)
		}
		return nil, fmt.Errorf("%w: %w", storage.ErrScanResult, err)
	}

	return results, nil
}

func (r *Repository) ListPvzByAcceptanceDateAndCity(
	ctx context.Context,
	pagination listparams.Pagination,
	startDate, endDate *time.Time,
) ([]PvzWithCityName, error) {
	sb := r.sqb.
		Select(
			"p.id as id",
			"p.registration_date as registration_date",
			"p.city_id as city_id",
			"c.name AS city_name",
		).
		From("pvz p").
		Join("receptions r ON r.pvz_id = p.id").
		Join("cities c ON c.id = p.city_id").
		OrderBy("p.registration_date DESC").
		Limit(uint64(pagination.Limit)).
		Offset(uint64(pagination.Offset()))

	if startDate != nil && endDate != nil {
		sb = sb.Where(
			sq.And{
				sq.GtOrEq{"r.date_time": startDate},
				sq.LtOrEq{"r.date_time": endDate},
			},
		)
	}

	sb = sb.GroupBy(
		"p.id",
		"p.registration_date",
		"p.city_id",
		"c.name",
	)
	sql, args, err := sb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrBuildQuery, err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrExecuteQuery, err)
	}
	defer rows.Close()

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[PvzWithCityName])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", storage.ErrNotFound, err)
		}
		return nil, fmt.Errorf("%w: %w", storage.ErrScanResult, err)
	}

	return results, nil
}

func (r *Repository) GetList(ctx context.Context) ([]Pvz, error) {
	selectBuilder := r.sqb.
		Select(Cols.ID, Cols.CityID, Cols.RegistrationDate).
		From((Pvz{}).TableName())

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrBuildQuery, err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrExecuteQuery, err)
	}
	defer rows.Close()

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[Pvz])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", storage.ErrNotFound, err)
		}
		return nil, fmt.Errorf("%w: %w", storage.ErrScanResult, err)
	}

	return results, nil
}
