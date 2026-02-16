package schema

import (
	"github.com/google/uuid"
	"github.com/valeragav/avito-pvz-service/internal/domain"
)

type ReceptionStatus struct {
	ID   uuid.UUID `db:"reception_statuses.id"`
	Name string    `db:"reception_statuses.name"`
}

func NewReceptionStatus(d *domain.ReceptionStatus) *ReceptionStatus {
	return &ReceptionStatus{
		ID:   d.ID,
		Name: string(d.Name),
	}
}

func NewDomainReceptionStatus(d *ReceptionStatus) *domain.ReceptionStatus {
	return &domain.ReceptionStatus{
		ID:   d.ID,
		Name: domain.ReceptionStatusCode(d.Name),
	}
}

func (ReceptionStatus) TableName() string {
	return "reception_statuses"
}

func (p ReceptionStatus) InsertColumns() []string {
	return []string{"id", "name"}
}

func (p ReceptionStatus) Columns() []string {
	return []string{"reception_statuses.id as \"reception_statuses.id\"", "reception_statuses.name as \"reception_statuses.name\""}
}

func (p ReceptionStatus) Values() []any {
	return []any{p.ID, p.Name}
}

var ReceptionStatusCols = struct {
	ID   string
	Name string
}{
	"id",
	"name",
}
