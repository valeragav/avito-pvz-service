package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type PVZ struct {
	ID               uuid.UUID
	RegistrationDate time.Time
	CityID           uuid.UUID

	Receptions []*Reception
	City       *City
}

var ErrPVZNotFound = errors.New("not found pvz")
var ErrDuplicatePvzID = errors.New("duplicate pvz id")
