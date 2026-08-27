package model_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	pkgerrs "github.com/maket12/ads-service/backend/authservice/pkg/errs"
	"github.com/maket12/ads-service/backend/searchservice/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAdIndex(t *testing.T) {
	var (
		testID        = gofakeit.UUID()
		testTitle     = gofakeit.ProductName()
		testDesc      = gofakeit.ProductDescription()
		testPrice     = int64(gofakeit.Uint64())
		testCategory  = model.CategoryElectronics.String()
		testMainImage = gofakeit.URL()
	)

	type testCase struct {
		name        string
		id          string
		title       string
		description string
		price       int64
		rawCategory string
		mainImage   string
		expect      error
	}

	var tests = []testCase{
		{
			name:        "success",
			id:          testID,
			title:       testTitle,
			description: testDesc,
			price:       testPrice,
			rawCategory: testCategory,
			mainImage:   testMainImage,
			expect:      nil,
		},
		{
			name:   "empty id",
			id:     "",
			expect: pkgerrs.ErrValueIsRequired,
		},
		{
			name:   "empty title",
			id:     testID,
			title:  "",
			expect: pkgerrs.ErrValueIsRequired,
		},
		{
			name:   "negative price",
			id:     testID,
			title:  testTitle,
			price:  -1 * testPrice,
			expect: pkgerrs.ErrValueIsInvalid,
		},
		{
			name:        "unsupported category",
			id:          testID,
			title:       testTitle,
			price:       testPrice,
			rawCategory: "unknown",
			expect:      pkgerrs.ErrValueIsInvalid,
		},
		{
			name:        "empty main image",
			id:          testID,
			title:       testTitle,
			price:       testPrice,
			rawCategory: testCategory,
			mainImage:   "",
			expect:      pkgerrs.ErrValueIsRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adIndex, err := model.NewAdIndex(
				tt.id, tt.title,
				tt.description, tt.price,
				tt.rawCategory, tt.mainImage,
			)
			if tt.expect == nil {
				require.NoError(t, err)
				assert.Equal(t, tt.id, adIndex.ID())
				assert.Equal(t, tt.title, adIndex.Title())
				assert.Equal(t, tt.description, adIndex.Description())
				assert.Equal(t, tt.price, adIndex.Price())
				assert.Equal(t, adIndex.Category().String(), tt.rawCategory)
				assert.Equal(t, tt.mainImage, adIndex.MainImage())
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expect)
				assert.Nil(t, adIndex)
			}
		})
	}
}
