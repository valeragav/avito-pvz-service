package repo

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/infra/repo/schema"
)

type ProductTypeRepository struct {
	db  *pgxpool.Pool
	sqb sq.StatementBuilderType
}

func NewProductTypeRepository(db *pgxpool.Pool) *ProductTypeRepository {
	return &ProductTypeRepository{
		db:  db,
		sqb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *ProductTypeRepository) Get(ctx context.Context, filter domain.ProductType) (*domain.ProductType, error) {
	where := sq.Eq{}
	if filter.ID != uuid.Nil {
		where[schema.ProductTypeCols.ID] = filter.ID
	}
	if filter.Name != "" {
		where[schema.ProductTypeCols.Name] = filter.Name
	}

	record := schema.NewProductType(&filter)

	qb := r.sqb.
		Select(record.Columns()...).
		From(record.TableName()).
		Where(where)

	result, err := CollectOneRow(ctx, r.db, qb, pgx.RowToStructByName[schema.ProductType])
	if err != nil {
		return nil, err
	}

	return schema.NewDomainProductType(&result), nil
}

func (r ProductTypeRepository) CreateBatch(ctx context.Context, productTypes []domain.ProductType) error {
	qb := r.sqb.
		Insert(schema.ProductType{}.TableName()).
		Columns(schema.ProductType{}.InsertColumns()...)

	for _, productType := range productTypes {
		if productType.ID == uuid.Nil {
			productType.ID = uuid.New()
		}

		qb = qb.Values(productType.ID, productType.Name)
	}

	_, err := CollectRows(ctx, r.db, qb, pgx.RowToStructByName[*schema.ProductType])
	if err != nil {
		return err
	}

	return nil
}
