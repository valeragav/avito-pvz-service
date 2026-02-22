package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID
	DateTime    time.Time
	TypeID      uuid.UUID
	ReceptionID uuid.UUID

	ProductType *ProductType
}

type ProductType struct {
	ID   uuid.UUID
	Name string
}

var ErrProductToDelete = errors.New("no products to delete")
