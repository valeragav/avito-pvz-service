package schema

import (
	"github.com/google/uuid"
	"github.com/valeragav/avito-pvz-service/internal/domain"
)

type ProductType struct {
	ID   uuid.UUID `db:"product_types.id"`
	Name string    `db:"product_types.name"`
}

func NewProductType(d *domain.ProductType) *ProductType {
	return &ProductType{
		ID:   d.ID,
		Name: d.Name,
	}
}

func NewDomainProductType(d *ProductType) *domain.ProductType {
	return &domain.ProductType{
		ID:   d.ID,
		Name: d.Name,
	}
}

func NewDomainProductTypeList(d []*ProductType) []*domain.ProductType {
	var res = make([]*domain.ProductType, 0, len(d))
	for _, record := range d {
		res = append(res, NewDomainProductType(record))
	}
	return res
}

func (ProductType) TableName() string {
	return "product_types"
}

func (p ProductType) InsertColumns() []string {
	return []string{"id", "name"}
}

func (p ProductType) Columns() []string {
	return []string{"product_types.id as \"product_types.id\"", "product_types.name as \"product_types.name\""}
}

func (p ProductType) Values() []any {
	return []any{p.ID, p.Name}
}

var ProductTypeCols = struct {
	ID   string
	Name string
}{
	"id",
	"name",
}
