package domain

import (
	"errors"

	"github.com/google/uuid"
)

type City struct {
	ID   uuid.UUID
	Name string
}

var ErrCityNotFound = errors.New("not found city")
