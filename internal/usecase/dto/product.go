package dto

import (
	"github.com/google/uuid"
)

type ProductCreate struct {
	TypeName string
	PvzID    uuid.UUID
}
