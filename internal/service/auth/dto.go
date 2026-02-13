package auth

import (
	"slices"
	"strings"

	"github.com/valeragav/avito-pvz-service/internal/infrastructure/storage/user"
)

type UserRole string

const (
	EmployeeRole  UserRole = "employee"
	ModeratorRole UserRole = "moderator"
)

func HasRequiredRole(userRole UserRole, requiredRoles []UserRole) bool {
	role := UserRole(strings.ToLower(string(userRole)))
	return slices.Contains(requiredRoles, role)
}

type RegisterIn struct {
	Email    string
	Password string
	Role     UserRole
}

type LoginIn struct {
	Email    string
	Password string
}

type RegisterOut struct {
	User user.User
}

func ToEnt(in RegisterIn) user.User {
	return user.User{
		Email: in.Email,
		Role:  string(in.Role),
	}
}
