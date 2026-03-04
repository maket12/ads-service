package dto

import "github.com/google/uuid"

type UpdateAdInput struct {
	AdID        uuid.UUID
	SellerID    uuid.UUID
	Title       *string
	Description *string
	Price       *int64
	Images      []string
}

type UpdateAdOutput struct {
	Success bool
}
