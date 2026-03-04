package dto

import "github.com/google/uuid"

type RejectAdInput struct {
	AdID     uuid.UUID
	SellerID uuid.UUID
}

type RejectAdOutput struct {
	Success bool
}
