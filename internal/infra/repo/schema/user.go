package schema

import (
	"github.com/google/uuid"
	"github.com/valeragav/avito-pvz-service/internal/domain"
)

type User struct {
	ID           uuid.UUID `db:"users.id"`
	Email        string    `db:"users.email"`
	PasswordHash string    `db:"users.password_hash"`
	Role         string    `db:"users.role"`
}

func NewUser(d *domain.User) *User {
	return &User{
		ID:           d.ID,
		Email:        d.Email,
		PasswordHash: d.PasswordHash,
		Role:         string(d.Role),
	}
}

func NewDomainUser(d *User) *domain.User {
	return &domain.User{
		ID:           d.ID,
		Email:        d.Email,
		PasswordHash: d.PasswordHash,
		Role:         domain.Role(d.Role),
	}
}

func (User) TableName() string {
	return "users"
}

func (u User) InsertColumns() []string {
	return []string{"id", "email", "password_hash", "role"}
}

func (u User) Columns() []string {
	return []string{"users.id as \"users.id\"", "users.email as \"users.email\"",
		"users.password_hash as \"users.password_hash\"", "users.role as \"users.role\""}
}

func (u User) Values() []any {
	return []any{u.ID, u.Email, u.PasswordHash, u.Role}
}

var UserCols = struct {
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
