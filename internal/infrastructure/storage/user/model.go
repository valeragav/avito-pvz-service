package user

import "github.com/google/uuid"

type User struct {
	ID           uuid.UUID `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	Role         string    `db:"role"`
}

func (User) TableName() string {
	return "users"
}

func (User) AllCols() []string {
	return []string{
		userCols.ID,
		userCols.Email,
		userCols.PasswordHash,
		userCols.Role,
	}
}

var userCols = struct {
	ID           string
	Email        string
	PasswordHash string
	Role         string
}{
	"id",
	"email",
	"password_hash",
	"role",
}
