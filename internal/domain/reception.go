package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type ReceptionStatusCode string

const (
	ReceptionStatusClose      ReceptionStatusCode = "close"
	ReceptionStatusInProgress ReceptionStatusCode = "in_progress"
)

type ReceptionStatus struct {
	ID   uuid.UUID
	Name ReceptionStatusCode
}

type Reception struct {
	ID       uuid.UUID
	PvzID    uuid.UUID
	DateTime time.Time
	StatusID uuid.UUID

	Products        []*Product
	ReceptionStatus *ReceptionStatus
}

var ErrNoReceptionIsCurrentlyInProgress = errors.New("no reception is currently in progress")
var ErrReceptionNotFound = errors.New("reception not found")
