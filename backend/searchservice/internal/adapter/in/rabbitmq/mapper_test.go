package rabbitmq_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/maket12/ads-service/backend/searchservice/internal/adapter/in/rabbitmq"
	"github.com/maket12/ads-service/backend/searchservice/internal/app/dto"

	"github.com/stretchr/testify/assert"
)

func TestMapAdPublishedEventToInput(t *testing.T) {
	t.Run("Success - maps all fields correctly", func(t *testing.T) {
		event := rabbitmq.AdPublishedEvent{
			ID:          gofakeit.UUID(),
			Title:       gofakeit.ProductName(),
			Description: gofakeit.ProductDescription(),
			Price:       gofakeit.Int64(),
			Category:    gofakeit.ProductCategory(),
			MainImage:   gofakeit.URL(),
		}

		out := rabbitmq.MapAdPublishedEventToInput(event)

		expected := dto.CreateAdIndexInput{
			ID:          event.ID,
			Title:       event.Title,
			Description: event.Description,
			Price:       event.Price,
			Category:    event.Category,
			MainImage:   event.MainImage,
		}

		assert.Equal(t, expected, out)
	})

	t.Run("Success - zero value event maps to zero value input", func(t *testing.T) {
		event := rabbitmq.AdPublishedEvent{}

		out := rabbitmq.MapAdPublishedEventToInput(event)

		assert.Equal(t, dto.CreateAdIndexInput{}, out)
	})

	t.Run("Success - field values are copied as-is (fuzz)", func(t *testing.T) {
		for i := 0; i < 20; i++ {
			event := rabbitmq.AdPublishedEvent{
				ID:          gofakeit.UUID(),
				Title:       gofakeit.Sentence(3),
				Description: gofakeit.Paragraph(1, 3, 10, " "),
				Price:       gofakeit.Int64(),
				Category:    gofakeit.Word(),
				MainImage:   gofakeit.URL(),
			}

			out := rabbitmq.MapAdPublishedEventToInput(event)

			assert.Equal(t, event.ID, out.ID)
			assert.Equal(t, event.Title, out.Title)
			assert.Equal(t, event.Description, out.Description)
			assert.Equal(t, event.Price, out.Price)
			assert.Equal(t, event.Category, out.Category)
			assert.Equal(t, event.MainImage, out.MainImage)
		}
	})
}
