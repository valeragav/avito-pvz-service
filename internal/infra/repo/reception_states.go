package repo

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/infra"
	"github.com/valeragav/avito-pvz-service/internal/infra/repo/schema"
)

type ReceptionStatusRepository struct {
	db  infra.DBTX
	sqb sq.StatementBuilderType
}

func NewReceptionStatusRepository(db infra.DBTX) *ReceptionStatusRepository {
	return &ReceptionStatusRepository{
		db:  db,
		sqb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *ReceptionStatusRepository) Get(ctx context.Context, filter domain.ReceptionStatus) (*domain.ReceptionStatus, error) {
	where := sq.Eq{}
	if filter.ID != uuid.Nil {
		where[schema.ReceptionStatusCols.ID] = filter.ID
	}
	if filter.Name != "" {
		where[schema.ReceptionStatusCols.Name] = filter.Name
	}

	record := schema.NewReceptionStatus(&filter)

	qb := r.sqb.
		Select(record.Columns()...).
		From(record.TableName()).
		Where(where)

	result, err := CollectOneRow(ctx, r.db, qb, pgx.RowToStructByName[schema.ReceptionStatus])
	if err != nil {
		return nil, err
	}

	return schema.NewDomainReceptionStatus(result), nil
}

func (r ReceptionStatusRepository) CreateBatch(ctx context.Context, statuses []domain.ReceptionStatus) error {
	qb := r.sqb.
		Insert(schema.ReceptionStatus{}.TableName()).
		Columns(schema.ReceptionStatus{}.InsertColumns()...)

	for _, status := range statuses {
		if status.ID == uuid.Nil {
			status.ID = uuid.New()
		}

		qb = qb.Values(status.ID, status.Name)
	}

	qb = qb.Suffix("ON CONFLICT (name) DO NOTHING")

	return Exec(ctx, r.db, qb)
}
