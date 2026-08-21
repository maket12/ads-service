package dto

import (
	"github.com/google/uuid"
)

type GetAdInput struct {
	AdID     uuid.UUID
	SellerID uuid.UUID
}

type GetAdOutput Ad
