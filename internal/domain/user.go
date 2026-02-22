package domain

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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

func NewUser(email, password string, role Role) (*User, error) {
	if !role.IsValid() {
		return nil, ErrInvalidRole
	}

	hash, err := generateHashPass(password)
	if err != nil {
		return nil, err
	}

	return &User{
		Email:        email,
		PasswordHash: hash,
		Role:         role,
	}, nil
}

func (r Role) IsValid() bool {
	switch r {
	case EmployeeRole, ModeratorRole:
		return true
	default:
		return false
	}
}

func generateHashPass(reqPass string) (string, error) {
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(reqPass), bcrypt.MinCost)
	if err != nil {
		return "", fmt.Errorf("generate password hash: %w", err)
	}
	return string(hashBytes), nil
}

var ErrInvalidEmailOrPassword = errors.New("invalid email or password")
var ErrAlreadyExists = errors.New("already exists")
var ErrInvalidRole = errors.New("invalid role")
