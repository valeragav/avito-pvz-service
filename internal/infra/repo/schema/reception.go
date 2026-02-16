package schema

import (
	"time"

	"github.com/google/uuid"
	"github.com/valeragav/avito-pvz-service/internal/domain"
)

type Reception struct {
	ID       uuid.UUID `db:"receptions.id"`
	DateTime time.Time `db:"receptions.date_time"`
	PvzID    uuid.UUID `db:"receptions.pvz_id"`
	StatusID uuid.UUID `db:"receptions.status_id"`
}

type ReceptionWithStatus struct {
	Reception
	ReceptionStatus
}

func NewReception(d *domain.Reception) *Reception {
	return &Reception{
		ID:       d.ID,
		DateTime: d.DateTime,
		PvzID:    d.PvzID,
		StatusID: d.StatusID,
	}
}

func NewDomainReception(d *Reception) *domain.Reception {
	return &domain.Reception{
		ID:       d.ID,
		DateTime: d.DateTime,
		PvzID:    d.PvzID,
		StatusID: d.StatusID,
	}
}

func NewDomainReceptionList(d []*Reception) []*domain.Reception {
	var res = make([]*domain.Reception, 0, len(d))
	for _, record := range d {
		res = append(res, NewDomainReception(record))
	}
	return res
}

func NewDomainReceptionWithStatus(d *ReceptionWithStatus) *domain.Reception {
	return &domain.Reception{
		ID:       d.Reception.ID,
		PvzID:    d.Reception.PvzID,
		DateTime: d.Reception.DateTime,
		StatusID: d.Reception.StatusID,
		ReceptionStatus: &domain.ReceptionStatus{
			ID:   d.ReceptionStatus.ID,
			Name: domain.ReceptionStatusCode(d.ReceptionStatus.Name),
		},
	}
}

func NewDomainReceptionWithStatusList(d []*ReceptionWithStatus) []*domain.Reception {
	var res = make([]*domain.Reception, 0, len(d))
	for _, record := range d {
		res = append(res, NewDomainReceptionWithStatus(record))
	}
	return res
}

func (p ReceptionWithStatus) Columns() []string {
	res := Reception{}.Columns()
	res = append(res, ReceptionStatus{}.Columns()...)
	return res
}

func (Reception) TableName() string {
	return "receptions"
}

func (p Reception) InsertColumns() []string {
	return []string{"id", "pvz_id", "status_id", "date_time"}
}

func (p Reception) Columns() []string {
	return []string{"receptions.id as \"receptions.id\"", "receptions.pvz_id as \"receptions.pvz_id\"",
		"receptions.status_id as \"receptions.status_id\"", "receptions.date_time as \"receptions.date_time\""}
}

func (p Reception) Values() []any {
	return []any{p.ID, p.PvzID, p.StatusID, p.DateTime}
}

var ReceptionCols = struct {
	ID       string
	DateTime string
	PvzID    string
	StatusID string
}{
	"id",
	"date_time",
	"pvz_id",
	"status_id",
}
