package receptions

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type CreateIn struct {
	PvzID uuid.UUID
}

type CreateOut struct {
	ID       uuid.UUID
	DateTime time.Time
	PvzID    uuid.UUID
	Status   string
}

type CloseLastReceptionOut struct {
	ID       uuid.UUID
	DateTime time.Time
	PvzID    uuid.UUID
	Status   string
}

var ErrNotFoundReceptionsRepoInProgress = errors.New("reception is not in progress")
