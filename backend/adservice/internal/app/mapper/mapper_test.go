package mapper_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/maket12/ads-service/backend/adservice/v2/internal/app/mapper"
	"github.com/maket12/ads-service/backend/adservice/v2/internal/domain/model"
	"github.com/maket12/ads-service/backend/adservice/v2/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestMapDomainToGetAdOut(t *testing.T) {
	ad := model.RestoreAd(
		uuid.New(),
		uuid.New(),
		"Test Title",
		utils.VPtr("Full description"),
		1000,
		model.CategoryVideoGames,
		model.AdPublished,
		[]string{"image1.jpg"},
		time.Now(),
		utils.VPtr(time.Now().Add(time.Second)),
	)

	result := mapper.MapDomainToGetAdOut(ad)

	assert.Equal(t, ad.ID(), result.AdID)
	assert.Equal(t, ad.SellerID(), result.SellerID)
	assert.Equal(t, ad.Title(), result.Title)
	assert.Equal(t, ad.Description(), result.Description)
	assert.Equal(t, ad.Price(), result.Price)
	assert.Equal(t, ad.Status().String(), result.Status)
	assert.Equal(t, ad.Images(), result.Images)
	assert.Equal(t, ad.CreatedAt(), result.CreatedAt)
	assert.Equal(t, ad.UpdatedAt(), result.UpdatedAt)
}

func TestMapDomainToListSellerAdsOut(t *testing.T) {
	ad := model.RestoreAd(
		uuid.New(),
		uuid.New(),
		"Test Title",
		nil,
		1000,
		model.CategoryVideoGames,
		model.AdPublished,
		[]string{"image1.jpg"},
		time.Now(),
		nil,
	)
	ads := []*model.Ad{ad}

	result := mapper.MapDomainToListSellerAdsOut(ads)

	assert.Len(t, result.Ads, 1)
	assert.Equal(t, ad.ID(), result.Ads[0].AdID)
	assert.Equal(t, ad.Title(), result.Ads[0].Title)
}

func TestMapDomainToListAdsOut(t *testing.T) {
	ad := model.RestoreAd(
		uuid.New(),
		uuid.New(),
		"Test Title",
		nil,
		1000,
		model.CategoryVideoGames,
		model.AdPublished,
		[]string{"image1.jpg"},
		time.Now(),
		nil,
	)
	ads := []*model.Ad{ad}

	result := mapper.MapDomainToListAdsOut(ads)

	assert.Len(t, result.Ads, 1)
	assert.Equal(t, ad.ID(), result.Ads[0].AdID)
	assert.Equal(t, ad.Title(), result.Ads[0].Title)
}
