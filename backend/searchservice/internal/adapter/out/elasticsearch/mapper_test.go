package elasticsearch

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/maket12/ads-service/backend/searchservice/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapAdIndexToEsDTO(t *testing.T) {
	var (
		id          = gofakeit.UUID()
		title       = gofakeit.ProductName()
		description = gofakeit.ProductDescription()
		price       = int64(gofakeit.Price(100, 10000))
		rawCategory = model.CategoryElectronics.String()
		mainImage   = gofakeit.URL()
	)

	adIndex, err := model.NewAdIndex(id, title, description, price, rawCategory, mainImage)
	require.NoError(t, err)

	dto := mapAdIndexToEsDTO(adIndex)

	assert.Equal(t, adIndex.ID(), dto.ID)
	assert.Equal(t, adIndex.Title(), dto.Title)
	assert.Equal(t, adIndex.Description(), dto.Description)
	assert.Equal(t, adIndex.Price(), dto.Price)
	assert.Equal(t, adIndex.Category().String(), dto.Category)
	assert.Equal(t, adIndex.MainImage(), dto.MainImage)
}

func TestMapEsDTOToAdIndex(t *testing.T) {
	dto := esAdDTO{
		ID:          gofakeit.UUID(),
		Title:       gofakeit.ProductName(),
		Description: gofakeit.ProductDescription(),
		Price:       int64(gofakeit.Price(100, 10000)),
		Category:    model.CategoryVehicles.String(),
		MainImage:   gofakeit.URL(),
	}

	adIndex := mapEsDTOToAdIndex(dto)
	require.NotNil(t, adIndex)

	assert.Equal(t, dto.ID, adIndex.ID())
	assert.Equal(t, dto.Title, adIndex.Title())
	assert.Equal(t, dto.Description, adIndex.Description())
	assert.Equal(t, dto.Price, adIndex.Price())
	assert.Equal(t, dto.Category, adIndex.Category().String())
	assert.Equal(t, dto.MainImage, adIndex.MainImage())
}
