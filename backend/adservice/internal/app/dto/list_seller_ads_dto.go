package dto

import (
	"github.com/google/uuid"
)

type ListSellerAdsInput struct {
	SellerID uuid.UUID
}

type ListSellerAdsOutput struct {
	Ads []Ad
}
