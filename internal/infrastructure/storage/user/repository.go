package user

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/VaLeraGav/avito-pvz-service/internal/infrastructure/storage"
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

func (r Repository) Create(ctx context.Context, user User) (*User, error) {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	queryBuilder := r.sqb.
		Insert(user.TableName()).
		Columns(userCols.ID, userCols.Email, userCols.PasswordHash, userCols.Role).
		Values(user.ID, user.Email, user.PasswordHash, user.Role).
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

	userCreate, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[User])
	if err != nil {
		return nil, fmt.Errorf("%w: %w", storage.ErrScanResult, err)
	}

	return &userCreate, nil
}

func (r *Repository) Get(ctx context.Context, filter User) (*User, error) {
	where := sq.Eq{}
	if filter.ID != uuid.Nil {
		where[userCols.ID] = filter.ID
	}
	if filter.Email != "" {
		where[userCols.Email] = filter.Email
	}
	if filter.Role != "" {
		where[userCols.Role] = filter.Role
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

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, fmt.Errorf("%w: %w", storage.ErrScanResult, err)
	}

	return &result, nil
}
