package products

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type CreateIn struct {
	TypeName string
	PvzID    uuid.UUID
}

type CreateOut struct {
	ID          uuid.UUID
	TypeName    string
	ReceptionID uuid.UUID
	DateTime    time.Time
}

var ErrNotFound = errors.New("reception not found")
var ErrNotFoundReceptionsRepoInProgress = errors.New("reception is not in progress")
