package grpc

import (
	"github.com/maket12/ads-service/backend/searchservice/api/proto/generated/search_v1"
	"github.com/maket12/ads-service/backend/searchservice/internal/app/dto"
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

func mapAdIndexesDTOToPb(items []dto.AdIndexDTO) []*search_v1.AdIndex {
	mapped := make([]*search_v1.AdIndex, len(items))
	for i := range mapped {
		mapped[i] = &search_v1.AdIndex{
			Id:          items[i].ID,
			Title:       items[i].Title,
			Description: items[i].Description,
			Price:       items[i].Price,
			Category:    items[i].Category,
			MainImage:   items[i].MainImage,
		}
	}
	return mapped
}

func MapSearchAdsDTOToPb(out dto.SearchAdsOutput) *search_v1.SearchAdsResponse {
	return &search_v1.SearchAdsResponse{
		Items: mapAdIndexesDTOToPb(out.Items),
		Total: out.Total,
	}
}
