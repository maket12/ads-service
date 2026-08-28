package grpc

import (
	"github.com/maket12/ads-service/backend/searchservice/internal/app/dto"
	"github.com/maket12/ads-service/backend/searchservice/pkg/generated/search_v1"
)

func MapSearchAdsPbToDTO(req *search_v1.SearchAdsRequest) dto.SearchAdsInput {
	return dto.SearchAdsInput{
		Text:      req.Text,
		Category:  req.Category,
		PriceFrom: req.PriceFrom,
		PriceTo:   req.PriceTo,
		Limit:     req.Limit,
		Offset:    req.Offset,
		SortBy:    req.SortBy,
	}
}
