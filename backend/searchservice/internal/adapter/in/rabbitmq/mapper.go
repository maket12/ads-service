package rabbitmq

import "github.com/maket12/ads-service/backend/searchservice/internal/app/dto"

func MapAdPublishedEventToInput(event AdPublishedEvent) dto.CreateAdIndexInput {
	return dto.CreateAdIndexInput{
		ID:          event.ID,
		Title:       event.Title,
		Description: event.Description,
		Price:       event.Price,
		Category:    event.Category,
		MainImage:   event.MainImage,
	}
}
