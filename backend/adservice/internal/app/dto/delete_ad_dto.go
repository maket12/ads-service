package dto

import "github.com/google/uuid"

type DeleteAdInput struct {
	AdID     uuid.UUID
	SellerID uuid.UUID
}

type DeleteAdOutput struct {
	Success bool
}
