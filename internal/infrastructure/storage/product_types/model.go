package product_types

import "github.com/google/uuid"

type ProductTypes struct {
	ID   uuid.UUID `db:"id"`
	Name string    `db:"name"`
}

func (ProductTypes) TableName() string {
	return "product_types"
}

func (ProductTypes) AllCols() []string {
	return []string{
		productTypeCols.ID,
		productTypeCols.Name,
	}
}

var productTypeCols = struct {
	ID   string
	Name string
}{
	"id",
	"name",
}
