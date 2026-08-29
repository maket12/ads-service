package rabbitmq_test

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/maket12/ads-service/backend/adservice/internal/adapter/out/rabbitmq"
	"github.com/maket12/ads-service/backend/adservice/internal/domain/model"
	"github.com/maket12/ads-service/backend/authservice/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestMapAdToAdPublishedEvent_FullData(t *testing.T) {
	ad, _ := model.NewAd(uuid.New(),
		gofakeit.ProductName(),
		utils.VPtr(gofakeit.ProductDescription()),
		int64(gofakeit.Uint16()),
		model.CategoryVideoGames.String(),
		[]string{gofakeit.URL(), gofakeit.URL()},
	)

	event := rabbitmq.MapAdToAdPublishedEvent(ad)

	assert.Equal(t, ad.ID().String(), event.ID)
	assert.Equal(t, ad.Title(), event.Title)
	assert.Equal(t, *ad.Description(), event.Description)
	assert.Equal(t, ad.Price(), event.Price)
	assert.Equal(t, ad.Category().String(), event.Category)
	assert.Equal(t, ad.Images()[0], event.MainImage)
}

func TestMapAdToAdPublishedEvent_NilAndEmptyFields(t *testing.T) {
	ad := model.RestoreAd(
		uuid.New(), uuid.New(),
		gofakeit.ProductName(),
		nil,
		int64(gofakeit.Uint16()),
		model.CategoryVideoGames,
		model.AdPublished,
		[]string{},
		time.Now(),
		nil,
	)

	event := rabbitmq.MapAdToAdPublishedEvent(ad)

	assert.Equal(t, ad.ID().String(), event.ID)
	assert.Equal(t, ad.Title(), event.Title)
	assert.Empty(t, event.Description)
	assert.Empty(t, event.MainImage)
}

func TestMapAdToAdUpdatedEvent_FullData(t *testing.T) {
	ad, _ := model.NewAd(uuid.New(),
		gofakeit.ProductName(),
		utils.VPtr(gofakeit.ProductDescription()),
		int64(gofakeit.Uint16()),
		model.CategoryVideoGames.String(),
		[]string{gofakeit.URL(), gofakeit.URL()},
	)

	event := rabbitmq.MapAdToAdUpdatedEvent(ad)

	assert.Equal(t, ad.ID().String(), event.ID)
	assert.Equal(t, ad.Title(), event.Title)
	assert.Equal(t, *ad.Description(), event.Description)
	assert.Equal(t, ad.Price(), event.Price)
	assert.Equal(t, ad.Category().String(), event.Category)
	assert.Equal(t, ad.Images()[0], event.MainImage)
}

func TestMapAdToAdUpdatedEvent_EmptyImagesSlice(t *testing.T) {
	ad := model.RestoreAd(
		uuid.New(), uuid.New(),
		gofakeit.ProductName(),
		nil,
		int64(gofakeit.Uint16()),
		model.CategoryVideoGames,
		model.AdPublished,
		[]string{},
		time.Now(),
		nil,
	)

	event := rabbitmq.MapAdToAdUpdatedEvent(ad)

	assert.Equal(t, ad.ID().String(), event.ID)
	assert.Empty(t, event.Description)
	assert.Empty(t, event.MainImage)
}
