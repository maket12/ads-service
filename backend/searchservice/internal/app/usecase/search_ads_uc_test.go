package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/maket12/ads-service/backend/adservice/pkg/utils"
	"github.com/maket12/ads-service/backend/searchservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/backend/searchservice/internal/app/errs"
	"github.com/maket12/ads-service/backend/searchservice/internal/app/usecase"
	"github.com/maket12/ads-service/backend/searchservice/internal/domain/model"
	"github.com/maket12/ads-service/backend/searchservice/internal/domain/port/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSearchAdsUC_Execute(t *testing.T) {
	type testCase struct {
		name          string
		input         dto.SearchAdsInput
		mockBehaviour func(a *mocks.MockAdIndexRepository)
		expectErr     error
		expectTotal   int64
	}

	text := gofakeit.ProductName()
	category := utils.VPtr(model.CategoryHome.String())
	priceFrom := utils.VPtr(int64(10))
	priceTo := utils.VPtr(int64(1000))
	limit := int32(10)
	offset := int32(0)
	sortBy := model.SortByPriceAsc.String()

	items := []*model.AdIndex{}
	total := int64(0)

	var tests = []testCase{
		{
			name: "Success",
			input: dto.SearchAdsInput{
				Text:      text,
				Category:  category,
				PriceFrom: priceFrom,
				PriceTo:   priceTo,
				Limit:     limit,
				Offset:    offset,
				SortBy:    sortBy,
			},
			mockBehaviour: func(a *mocks.MockAdIndexRepository) {
				a.EXPECT().
					Search(mock.Anything, mock.AnythingOfType("*model.SearchQuery")).
					Return(items, total, nil)
			},
			expectErr:   nil,
			expectTotal: total,
		},
		{
			name: "Failure - invalid input",
			input: dto.SearchAdsInput{
				Text:      text,
				Category:  category,
				PriceFrom: priceTo,
				PriceTo:   priceFrom,
				Limit:     limit,
				Offset:    offset,
				SortBy:    sortBy,
			},
			mockBehaviour: func(_ *mocks.MockAdIndexRepository) {},
			expectErr:     ucerrs.ErrInvalidInput,
		},
		{
			name: "Failure - search es error",
			input: dto.SearchAdsInput{
				Text:      text,
				Category:  category,
				PriceFrom: priceFrom,
				PriceTo:   priceTo,
				Limit:     limit,
				Offset:    offset,
				SortBy:    sortBy,
			},
			mockBehaviour: func(a *mocks.MockAdIndexRepository) {
				a.EXPECT().
					Search(mock.Anything, mock.AnythingOfType("*model.SearchQuery")).
					Return(nil, int64(0), errors.New("es error"))
			},
			expectErr: ucerrs.ErrSearchAdIndexesES,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adIndexRepo := mocks.NewMockAdIndexRepository(t)

			tt.mockBehaviour(adIndexRepo)

			uc := usecase.NewSearchAdsUC(adIndexRepo)

			out, err := uc.Execute(context.Background(), tt.input)

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectTotal, out.Total)
			}
		})
	}
}
