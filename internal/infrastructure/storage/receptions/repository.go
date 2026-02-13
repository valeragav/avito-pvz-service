package receptions

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valeragav/avito-pvz-service/internal/infrastructure/storage"
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

func (r Repository) Create(ctx context.Context, reception Receptions) (*Receptions, error) {
	if reception.ID == uuid.Nil {
		reception.ID = uuid.New()
	}

	queryBuilder := r.sqb.
		Insert(reception.TableName()).
		Columns(Cols.ID, Cols.DateTime, Cols.PvzID, Cols.StatusID).
		Values(reception.ID, reception.DateTime, reception.PvzID, reception.StatusID).
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

	productCreate, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Receptions])
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrScanResult, err)
	}

	return &productCreate, nil
}

func (r *Repository) GetList(ctx context.Context, filter Receptions) ([]Receptions, error) {
	where := ToWhereMap(filter)

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

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[Receptions])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, fmt.Errorf("%w: %w", storage.ErrScanResult, err)
	}

	return results, nil
}

func (r *Repository) GetByIDs(ctx context.Context, receptionIDs []uuid.UUID) ([]Receptions, error) {
	selectBuilder := r.sqb.
		Select((Receptions{}).AllCols()...).
		From((Receptions{}).TableName()).
		Where(sq.Eq{Cols.PvzID: receptionIDs})

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrBuildQuery, err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrExecuteQuery, err)
	}
	defer rows.Close()

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[Receptions])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, fmt.Errorf("%w: %w", storage.ErrScanResult, err)
	}

	return results, nil
}

func (r *Repository) Get(ctx context.Context, filter Receptions) (*Receptions, error) {
	where := ToWhereMap(filter)

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

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Receptions])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, fmt.Errorf("%w: %w", storage.ErrScanResult, err)
	}

	return &result, nil
}

func (r *Repository) ListByIDsWithStatus(ctx context.Context, receptionIDs []uuid.UUID) ([]ReceptionsWithStatus, error) {
	// TODO: нужно ли делать запросом или лучше строит строки из entity
	selectBuilder := r.sqb.
		Select(
			"r.id as id",
			"r.date_time as date_time",
			"r.pvz_id as pvz_id",
			"r.status_id as status_id",
			"s.name AS status_name",
		).
		From((Receptions{}).TableName() + " r").
		Join("statuses s ON s.id = r.status_id").
		Where(sq.Eq{"r.pvz_id": receptionIDs})

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrBuildQuery, err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrExecuteQuery, err)
	}
	defer rows.Close()

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[ReceptionsWithStatus])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, fmt.Errorf("%w: %w", storage.ErrScanResult, err)
	}

	return results, nil
}

func (r *Repository) GetLastWithStatus(ctx context.Context, filter Receptions) (*ReceptionsWithStatus, error) {
	where := ToWhereMap(filter)

	selectBuilder := r.sqb.
		Select(
			"r.id as id",
			"r.date_time as date_time",
			"r.pvz_id as pvz_id",
			"r.status_id as status_id",
			"s.name AS status_name",
		).
		From((Receptions{}).TableName() + " r").
		Join("statuses s ON s.id = r.status_id").
		Where(where).
		OrderBy("r.date_time DESC").
		Limit(1)

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrBuildQuery, err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrExecuteQuery, err)
	}
	defer rows.Close()

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[ReceptionsWithStatus])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, fmt.Errorf("%w: %w", storage.ErrScanResult, err)
	}

	return &result, nil
}

func (r *Repository) Update(ctx context.Context, receptionID uuid.UUID, update Receptions) (*Receptions, error) {
	qb := r.sqb.
		Update((Receptions{}).TableName()).
		Where(sq.Eq{Cols.ID: receptionID}).
		Suffix("RETURNING *")

	var clauses = make(map[string]any)

	if update.ID != uuid.Nil {
		clauses[Cols.ID] = update.ID
	}
	if !update.DateTime.IsZero() {
		clauses[Cols.DateTime] = update.DateTime
	}
	if update.PvzID != uuid.Nil {
		clauses[Cols.PvzID] = update.PvzID
	}
	if update.StatusID != uuid.Nil {
		clauses[Cols.StatusID] = update.StatusID
	}

	qb = qb.SetMap(clauses)

	sql, args, err := qb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrBuildQuery, err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrExecuteQuery, err)
	}
	defer rows.Close()

	queryRes, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Receptions])
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrScanResult, err)
	}

	return &queryRes, nil
}

func ToWhereMap(elem Receptions) sq.Eq {
	where := sq.Eq{}

	if elem.ID != uuid.Nil {
		where[Cols.ID] = elem.ID
	}
	if !elem.DateTime.IsZero() {
		where[Cols.DateTime] = elem.DateTime
	}
	if elem.PvzID != uuid.Nil {
		where[Cols.PvzID] = elem.PvzID
	}
	if elem.StatusID != uuid.Nil {
		where[Cols.StatusID] = elem.StatusID
	}
	return where
}
