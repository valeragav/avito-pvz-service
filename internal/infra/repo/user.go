package repo

import (
	"context"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/infra"
	"github.com/valeragav/avito-pvz-service/internal/infra/repo/schema"
)

type UserRepository struct {
	db  infra.DBTX
	sqb sq.StatementBuilderType
}

func NewUserRepository(db infra.DBTX) *UserRepository {
	return &UserRepository{
		db:  db,
		sqb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r UserRepository) Create(ctx context.Context, user domain.User) (*domain.User, error) {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	record := schema.NewUser(&user)

	qb := r.sqb.
		Insert(record.TableName()).
		Columns(record.InsertColumns()...).
		Values(record.Values()...).
		Suffix(fmt.Sprintf("RETURNING %s", strings.Join(record.Columns(), ", ")))

	userCreate, err := CollectOneRow(ctx, r.db, qb, pgx.RowToStructByName[schema.User])
	if err != nil {
		return nil, err
	}

	return schema.NewDomainUser(&userCreate), nil
}

func (r *UserRepository) Get(ctx context.Context, filter domain.User) (*domain.User, error) {
	where := sq.Eq{}
	if filter.ID != uuid.Nil {
		where[schema.UserCols.ID] = filter.ID
	}
	if filter.Email != "" {
		where[schema.UserCols.Email] = filter.Email
	}
	if filter.Role != "" {
		where[schema.UserCols.Role] = filter.Role
	}

	record := schema.NewUser(&filter)

	qb := r.sqb.
		Select(record.Columns()...).
		From(record.TableName()).
		Where(where)

	result, err := CollectOneRow(ctx, r.db, qb, pgx.RowToStructByName[schema.User])
	if err != nil {
		return nil, err
	}

	return schema.NewDomainUser(&result), nil
}
