package mapper

import (
	"github.com/maket12/ads-service/adservice/internal/app/dto"
	"github.com/maket12/ads-service/adservice/internal/domain/model"
)

func mapDomainToAdDTO(ad *model.Ad) dto.Ad {
	return dto.Ad{
		AdID:        ad.ID(),
		SellerID:    ad.SellerID(),
		Title:       ad.Title(),
		Description: ad.Description(),
		Price:       ad.Price(),
		Status:      ad.Status().String(),
		Images:      ad.Images(),
		CreatedAt:   ad.CreatedAt(),
		UpdatedAt:   ad.UpdatedAt(),
	}
}

func MapDomainToGetAdOut(ad *model.Ad) dto.GetAdOutput {
	return dto.GetAdOutput(mapDomainToAdDTO(ad))
}

func MapDomainToListSellerAdsOut(ads []*model.Ad) dto.ListSellerAdsOutput {
	mapped := make([]dto.Ad, len(ads))
	for i := range mapped {
		mapped[i] = mapDomainToAdDTO(ads[i])
	}
	return dto.ListSellerAdsOutput{Ads: mapped}
}

func MapDomainToListAdsOut(ads []*model.Ad) dto.ListAdsOutput {
	mapped := make([]dto.Ad, len(ads))
	for i := range mapped {
		mapped[i] = mapDomainToAdDTO(ads[i])
	}
	return dto.ListAdsOutput{Ads: mapped}
}
