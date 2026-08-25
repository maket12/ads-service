package dto

import (
	"time"

	"github.com/google/uuid"
)

type Ad struct {
	AdID        uuid.UUID
	SellerID    uuid.UUID
	Title       string
	Description *string
	Price       int64
	Category    string
	Status      string
	Images      []string
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}
