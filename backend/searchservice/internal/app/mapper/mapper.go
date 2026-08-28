package mapper

import (
	"github.com/maket12/ads-service/backend/searchservice/internal/app/dto"
	"github.com/maket12/ads-service/backend/searchservice/internal/domain/model"
)

func mapDomainAdIndexToDTO(adIndex *model.AdIndex) dto.AdIndexDTO {
	return dto.AdIndexDTO{
		ID:          adIndex.ID(),
		Title:       adIndex.Title(),
		Description: adIndex.Description(),
		Price:       adIndex.Price(),
		Category:    adIndex.Category().String(),
		MainImage:   adIndex.MainImage(),
	}
}

func MapDomainAdIndexesToDTO(adIndexes []*model.AdIndex) []dto.AdIndexDTO {
	mapped := make([]dto.AdIndexDTO, len(adIndexes))
	for i := range mapped {
		mapped[i] = mapDomainAdIndexToDTO(adIndexes[i])
	}
	return mapped
}
