package model_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/maket12/ads-service/backend/searchservice/internal/domain/model"
	pkgerrs "github.com/maket12/ads-service/backend/searchservice/pkg/errs"
	"github.com/maket12/ads-service/backend/searchservice/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSearchQuery(t *testing.T) {
	var (
		testText      = gofakeit.ProductName()
		testCategory  = utils.VPtr(model.CategoryElectronics.String())
		testPriceFrom = utils.VPtr(int64(2000))
		testPriceTo   = utils.VPtr(*testPriceFrom * 2)
		testLimit     = int32(50)
		testOffset    = int32(10)
		testSortBy    = model.SortByPriceAsc.String()
	)

	type testCase struct {
		name        string
		text        string
		rawCategory *string
		priceFrom   *int64
		priceTo     *int64
		limit       int32
		offset      int32
		sortBy      string
		expect      error
	}

	var tests = []testCase{
		{
			name:        "success",
			text:        testText,
			rawCategory: testCategory,
			priceFrom:   testPriceFrom,
			priceTo:     testPriceTo,
			limit:       testLimit,
			offset:      testOffset,
			sortBy:      testSortBy,
			expect:      nil,
		},
		{
			name:        "unsupported category",
			text:        testText,
			rawCategory: utils.VPtr("unknown"),
			expect:      pkgerrs.ErrValueIsInvalid,
		},
		{
			name:        "negative price from",
			text:        testText,
			rawCategory: testCategory,
			priceFrom:   utils.VPtr(int64(-100)),
			expect:      pkgerrs.ErrValueIsInvalid,
		},
		{
			name:        "negative price to",
			text:        testText,
			rawCategory: testCategory,
			priceFrom:   testPriceFrom,
			priceTo:     utils.VPtr(int64(-100)),
			expect:      pkgerrs.ErrValueIsInvalid,
		},
		{
			name:        "negative price range",
			text:        testText,
			rawCategory: testCategory,
			priceFrom:   utils.VPtr(int64(200)),
			priceTo:     utils.VPtr(int64(100)),
			expect:      pkgerrs.ErrValueIsInvalid,
		},
		{
			name:        "unsupported sort option",
			text:        testText,
			rawCategory: testCategory,
			priceFrom:   testPriceFrom,
			priceTo:     testPriceTo,
			sortBy:      "unknown",
			expect:      pkgerrs.ErrValueIsInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			searchQuery, err := model.NewSearchQuery(
				tt.text, tt.rawCategory,
				tt.priceFrom, tt.priceTo, tt.limit,
				tt.offset, tt.sortBy,
			)
			if tt.expect == nil {
				require.NoError(t, err)
				assert.Equal(t, tt.text, searchQuery.Text())
				assert.Equal(t, *tt.rawCategory, searchQuery.Category().String())
				assert.Equal(t, tt.priceFrom, searchQuery.PriceFrom())
				assert.Equal(t, tt.priceTo, searchQuery.PriceTo())
				assert.Equal(t, tt.limit, searchQuery.Limit())
				assert.Equal(t, tt.offset, searchQuery.Offset())
				assert.Equal(t, tt.sortBy, searchQuery.SortBy().String())
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expect)
				assert.Nil(t, searchQuery)
			}
		})
	}
}
