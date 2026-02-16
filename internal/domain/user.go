package domain

import (
	"errors"

	"github.com/google/uuid"
)

type Role string

const (
	ModeratorRole Role = "moderator"
	EmployeeRole  Role = "employee"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	Role         Role
}

type Token string

type UserClaims struct {
	Role Role
}

var ErrInvalidEmailOrPassword = errors.New("invalid email or password")
var ErrAlreadyExists = errors.New("already exists")
