package elasticsearch

import "github.com/maket12/ads-service/backend/searchservice/internal/domain/model"

func mapAdIndexToEsDTO(adIndex *model.AdIndex) esAdDTO {
	return esAdDTO{
		ID:          adIndex.ID(),
		Title:       adIndex.Title(),
		Description: adIndex.Description(),
		Price:       adIndex.Price(),
		Category:    adIndex.Category().String(),
		MainImage:   adIndex.MainImage(),
	}
}
