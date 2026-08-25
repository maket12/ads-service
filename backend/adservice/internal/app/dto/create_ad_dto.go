package dto

import "github.com/google/uuid"

type CreateAdInput struct {
	SellerID    uuid.UUID
	Title       string
	Description *string
	Price       int64
	Category    string
	Images      []string
}

type CreateAdOutput struct {
	AdID uuid.UUID
}
