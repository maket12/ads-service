package mapper_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/maket12/ads-service/backend/searchservice/internal/app/mapper"
	"github.com/maket12/ads-service/backend/searchservice/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapDomainAdIndexesToDTO(t *testing.T) {
	t.Run("success mapping non-empty slice", func(t *testing.T) {
		ad1, err := model.NewAdIndex(
			gofakeit.UUID(),
			gofakeit.ProductName(),
			gofakeit.ProductDescription(),
			int64(100),
			model.CategoryElectronics.String(),
			gofakeit.URL(),
		)
		require.NoError(t, err)

		ad2, err := model.NewAdIndex(
			gofakeit.UUID(),
			gofakeit.ProductName(),
			gofakeit.ProductDescription(),
			int64(200),
			model.CategoryVehicles.String(),
			gofakeit.URL(),
		)
		require.NoError(t, err)

		domainList := []*model.AdIndex{ad1, ad2}

		resultDTOs := mapper.MapDomainAdIndexesToDTO(domainList)

		require.Len(t, resultDTOs, len(domainList))

		assert.Equal(t, ad1.ID(), resultDTOs[0].ID)
		assert.Equal(t, ad1.Title(), resultDTOs[0].Title)
		assert.Equal(t, ad1.Description(), resultDTOs[0].Description)
		assert.Equal(t, ad1.Price(), resultDTOs[0].Price)
		assert.Equal(t, ad1.Category().String(), resultDTOs[0].Category)
		assert.Equal(t, ad1.MainImage(), resultDTOs[0].MainImage)

		assert.Equal(t, ad2.ID(), resultDTOs[1].ID)
		assert.Equal(t, ad2.Title(), resultDTOs[1].Title)
		assert.Equal(t, ad2.Description(), resultDTOs[1].Description)
		assert.Equal(t, ad2.Price(), resultDTOs[1].Price)
		assert.Equal(t, ad2.Category().String(), resultDTOs[1].Category)
		assert.Equal(t, ad2.MainImage(), resultDTOs[1].MainImage)
	})

	t.Run("empty slice returns empty result", func(t *testing.T) {
		var domainList []*model.AdIndex

		resultDTOs := mapper.MapDomainAdIndexesToDTO(domainList)

		assert.NotNil(t, resultDTOs)
		assert.Empty(t, resultDTOs)
	})
}
