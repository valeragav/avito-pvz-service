package cities

import "github.com/google/uuid"

type Cities struct {
	ID   uuid.UUID `db:"id"`
	Name string    `db:"name"`
}

func (Cities) TableName() string {
	return "cities"
}

func (Cities) AllCols() []string {
	return []string{
		cityCols.ID,
		cityCols.Name,
	}
}

var cityCols = struct {
	ID   string
	Name string
}{
	"id",
	"name",
}
