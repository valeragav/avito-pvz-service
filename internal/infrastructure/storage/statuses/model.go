package statuses

import "github.com/google/uuid"

var (
	StatusClose      = "close"
	StatusInProgress = "in_progress"
)

type Statuses struct {
	ID   uuid.UUID `db:"id"`
	Name string    `db:"name"`
}

func (Statuses) TableName() string {
	return "statuses"
}

func (Statuses) AllCols() []string {
	return []string{
		StatusCols.ID,
		StatusCols.Name,
	}
}

var StatusCols = struct {
	ID   string
	Name string
}{
	"id",
	"name",
}
