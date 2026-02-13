package receptions

import (
	"time"

	"github.com/google/uuid"
)

type ReceptionsWithStatus struct {
	ID         uuid.UUID `db:"id"`
	DateTime   time.Time `db:"date_time"`
	PvzID      uuid.UUID `db:"pvz_id"`
	StatusID   uuid.UUID `db:"status_id"`
	StatusName string    `db:"status_name"`
}

type Receptions struct {
	ID       uuid.UUID `db:"id"`
	DateTime time.Time `db:"date_time"`
	PvzID    uuid.UUID `db:"pvz_id"`
	StatusID uuid.UUID `db:"status_id"`
}

func (Receptions) TableName() string {
	return "receptions"
}

func (Receptions) AllCols() []string {
	return []string{
		Cols.ID,
		Cols.DateTime,
		Cols.PvzID,
		Cols.StatusID,
	}
}

func (r Receptions) AliasedCols(alias string) []string {
	cols := r.AllCols()

	out := make([]string, 0, len(cols))
	for _, c := range cols {
		out = append(out, alias+"."+c)
	}

	return out
}

var Cols = struct {
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
