package model_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	pkgerrs "github.com/maket12/ads-service/backend/authservice/pkg/errs"
	"github.com/maket12/ads-service/backend/authservice/pkg/utils"
	"github.com/maket12/ads-service/backend/searchservice/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAdIndex(t *testing.T) {
	var (
		testID        = gofakeit.UUID()
		testTitle     = gofakeit.ProductName()
		testDesc      = gofakeit.ProductDescription()
		testPrice     = int64(1000)
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

func TestAdIndex_Update(t *testing.T) {
	var (
		initialID        = gofakeit.UUID()
		initialTitle     = gofakeit.ProductName()
		initialDesc      = gofakeit.ProductDescription()
		initialPrice     = int64(1000)
		initialCategory  = model.CategoryElectronics.String()
		initialMainImage = gofakeit.URL()

		newTitle     = gofakeit.ProductName()
		newDesc      = gofakeit.ProductDescription()
		newPrice     = int64(2000)
		newCategory  = model.CategoryVehicles.String()
		newMainImage = gofakeit.URL()
	)

	type testCase struct {
		name        string
		title       *string
		description *string
		price       *int64
		rawCategory *string
		mainImage   *string
		expect      error
	}

	var tests = []testCase{
		{
			name:        "success",
			title:       utils.VPtr(newTitle),
			description: utils.VPtr(newDesc),
			price:       utils.VPtr(newPrice),
			rawCategory: utils.VPtr(newCategory),
			mainImage:   utils.VPtr(newMainImage),
			expect:      nil,
		},
		{
			name:   "empty title",
			title:  utils.VPtr(""),
			expect: pkgerrs.ErrValueIsInvalid,
		},
		{
			name:   "negative price",
			price:  utils.VPtr(-1 * newPrice),
			expect: pkgerrs.ErrValueIsInvalid,
		},
		{
			name:        "unsupported category",
			rawCategory: utils.VPtr("unknown"),
			expect:      pkgerrs.ErrValueIsInvalid,
		},
		{
			name:      "empty main image",
			mainImage: utils.VPtr(""),
			expect:    pkgerrs.ErrValueIsRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adIndex, err := model.NewAdIndex(
				initialID, initialTitle,
				initialDesc, initialPrice,
				initialCategory, initialMainImage,
			)
			require.NoError(t, err)

			err = adIndex.Update(
				tt.title, tt.description,
				tt.price, tt.rawCategory,
				tt.mainImage,
			)

			if tt.expect == nil {
				require.NoError(t, err)
				assert.Equal(t, initialID, adIndex.ID())
				assert.Equal(t, *tt.title, adIndex.Title())
				assert.Equal(t, *tt.description, adIndex.Description())
				assert.Equal(t, *tt.price, adIndex.Price())
				assert.Equal(t, adIndex.Category().String(), *tt.rawCategory)
				assert.Equal(t, *tt.mainImage, adIndex.MainImage())
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expect)
			}
		})
	}
}
