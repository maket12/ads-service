package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/maket12/ads-service/adservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/adservice/internal/app/errs"
	"github.com/maket12/ads-service/adservice/internal/app/usecase"
	"github.com/maket12/ads-service/adservice/internal/domain/model"
	"github.com/maket12/ads-service/adservice/internal/domain/port/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListAllAdsUC_Execute(t *testing.T) {
	type testCase struct {
		name          string
		input         dto.ListAllAdsInput
		mockBehaviour func(a *mocks.MockAdRepository)
		expectErr     error
	}

	limit, offset := 10, 0
	sellerID := uuid.New()

	var tests = []testCase{
		{
			name:  "Success",
			input: dto.ListAllAdsInput{Limit: limit, Offset: offset},
			mockBehaviour: func(a *mocks.MockAdRepository) {
				ad, _ := model.NewAd(sellerID, "title", nil, 100, nil)
				a.EXPECT().
					ListAds(mock.Anything, limit, offset).
					Return([]*model.Ad{ad}, nil)
			},
			expectErr: nil,
		},
		{
			name:  "Failure - db error",
			input: dto.ListAllAdsInput{Limit: limit, Offset: offset},
			mockBehaviour: func(a *mocks.MockAdRepository) {
				a.EXPECT().
					ListAds(mock.Anything, limit, offset).
					Return(nil, errors.New("db error"))
			},
			expectErr: ucerrs.ErrListAdsDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adRepo := mocks.NewMockAdRepository(t)
			tt.mockBehaviour(adRepo)

			uc := usecase.NewListAllAdsUC(adRepo)

			_, err := uc.Execute(context.Background(), tt.input)

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
