package dto

import "github.com/google/uuid"

type DeleteAllAdsInput struct {
	SellerID uuid.UUID
}

type DeleteAllAdsOutput struct {
	Success bool
}
