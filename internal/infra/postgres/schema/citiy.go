package schema

import (
	"github.com/google/uuid"
	"github.com/valeragav/avito-pvz-service/internal/domain"
)

type City struct {
	ID   uuid.UUID `db:"cities.id"`
	Name string    `db:"cities.name"`
}

func NewCity(d *domain.City) *City {
	return &City{
		ID:   d.ID,
		Name: d.Name,
	}
}

func NewDomainCities(d *City) *domain.City {
	return &domain.City{
		ID:   d.ID,
		Name: d.Name,
	}
}

func (City) TableName() string {
	return "cities"
}

func (c City) InsertColumns() []string {
	return []string{"id", "name"}
}

func (c City) Columns() []string {
	return []string{"cities.id as \"cities.id\"", "cities.name as \"cities.name\""}
}

func (c City) Values() []any {
	return []any{c.ID, c.Name}
}

var CityCols = struct {
	ID   string
	Name string
}{
	"id",
	"name",
}
