package dto

import (
	"time"

	"github.com/google/uuid"
)

type GetAdInput struct {
	AdID     uuid.UUID
	SellerID uuid.UUID
}

type GetAdOutput struct {
	AdID        uuid.UUID
	SellerID    uuid.UUID
	Title       string
	Description *string
	Price       int64
	Status      string
	Images      []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
