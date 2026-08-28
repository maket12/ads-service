package grpc_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/maket12/ads-service/backend/searchservice/api/proto/generated/search_v1"
	"github.com/maket12/ads-service/backend/searchservice/internal/adapter/in/grpc"
	"github.com/maket12/ads-service/backend/searchservice/internal/app/dto"
	"github.com/stretchr/testify/assert"
)

func TestMapSearchAdsPbToDTO(t *testing.T) {
	text := "laptop"
	category := "electronics"
	var priceFrom int64 = 500
	var priceTo int64 = 1500
	var limit int32 = 10
	var offset int32 = 0
	sortBy := "price_asc"

	req := &search_v1.SearchAdsRequest{
		Text:      text,
		Category:  &category,
		PriceFrom: &priceFrom,
		PriceTo:   &priceTo,
		Limit:     limit,
		Offset:    offset,
		SortBy:    sortBy,
	}

	result := grpc.MapSearchAdsPbToDTO(req)

	assert.Equal(t, req.Text, result.Text)
	assert.Equal(t, req.Category, result.Category)
	assert.Equal(t, req.PriceFrom, result.PriceFrom)
	assert.Equal(t, req.PriceTo, result.PriceTo)
	assert.Equal(t, req.Limit, result.Limit)
	assert.Equal(t, req.Offset, result.Offset)
	assert.Equal(t, req.SortBy, result.SortBy)
}

func TestMapSearchAdsDTOToPb(t *testing.T) {
	adID := uuid.New().String()
	out := dto.SearchAdsOutput{
		Items: []dto.AdIndexDTO{
			{
				ID:          adID,
				Title:       "Gaming Laptop",
				Description: "High performance laptop",
				Price:       1200,
				Category:    "electronics",
				MainImage:   "http://example.com/laptop.jpg",
			},
		},
		Total: 1,
	}

	result := grpc.MapSearchAdsDTOToPb(out)

	assert.Equal(t, out.Total, result.Total)
	assert.Len(t, result.Items, 1)

	assert.Equal(t, out.Items[0].ID, result.Items[0].Id)
	assert.Equal(t, out.Items[0].Title, result.Items[0].Title)
	assert.Equal(t, out.Items[0].Description, result.Items[0].Description)
	assert.Equal(t, out.Items[0].Price, result.Items[0].Price)
	assert.Equal(t, out.Items[0].Category, result.Items[0].Category)
	assert.Equal(t, out.Items[0].MainImage, result.Items[0].MainImage)
}
