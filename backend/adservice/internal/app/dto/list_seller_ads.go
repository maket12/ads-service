package dto

import (
	"time"

	"github.com/google/uuid"
)

type ListSellerAdsInput struct {
	SellerID uuid.UUID
	Limit    int
	Offset   int
}

type ListSellerAdsOutput struct {
	Ads []struct {
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
}
