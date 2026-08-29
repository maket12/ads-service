package rabbitmq

import "github.com/maket12/ads-service/backend/adservice/internal/domain/model"

func MapAdToAdPublishedEvent(ad *model.Ad) AdPublishedEvent {
	var desc, img string
	if ad.Description() != nil {
		desc = *ad.Description()
	}
	if ad.Images() != nil && len(ad.Images()) != 0 {
		img = ad.Images()[0]
	}

	return AdPublishedEvent{
		ID:          ad.ID().String(),
		Title:       ad.Title(),
		Description: desc,
		Price:       ad.Price(),
		Category:    ad.Category().String(),
		MainImage:   img,
	}
}

func MapAdToAdUpdatedEvent(ad *model.Ad) AdUpdatedEvent {
	var desc, img string
	if ad.Description() != nil {
		desc = *ad.Description()
	}
	if ad.Images() != nil && len(ad.Images()) != 0 {
		img = ad.Images()[0]
	}

	return AdUpdatedEvent{
		ID:          ad.ID().String(),
		Title:       ad.Title(),
		Description: desc,
		Price:       ad.Price(),
		Category:    ad.Category().String(),
		MainImage:   img,
	}
}
