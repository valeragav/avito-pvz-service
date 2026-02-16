package dto

import "github.com/valeragav/avito-pvz-service/internal/domain"

type RegisterIn struct {
	Email    string
	Password string
	Role     domain.Role
}

type LoginIn struct {
	Email    string
	Password string
}
