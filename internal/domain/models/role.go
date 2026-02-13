package user

// TODO: стоит ли перенести все model сюда
import (
	"slices"
	"strings"
)

type Role string

const (
	EmployeeRole  Role = "employee"
	ModeratorRole Role = "moderator"
)

func HasRequiredRole(userRole Role, requiredRoles []Role) bool {
	role := Role(strings.ToLower(string(userRole)))
	return slices.Contains(requiredRoles, role)
}
