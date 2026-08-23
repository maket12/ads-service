package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/maket12/ads-service/backend/adservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/backend/adservice/internal/app/errs"
	"github.com/maket12/ads-service/backend/adservice/internal/app/usecase"
	"github.com/maket12/ads-service/backend/adservice/internal/domain/model"
	"github.com/maket12/ads-service/backend/adservice/internal/domain/port/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListSellerAdsUC_Execute(t *testing.T) {
	type testCase struct {
		name          string
		mockBehaviour func(a *mocks.MockAdRepository, sellerID uuid.UUID)
		expectErr     error
	}

	sellerID := uuid.New()

	var tests = []testCase{
		{
			name: "Success",
			mockBehaviour: func(a *mocks.MockAdRepository, sellerID uuid.UUID) {
				ad, _ := model.NewAd(sellerID, "title", nil, 100, nil)
				a.EXPECT().
					ListSellerAds(mock.Anything, sellerID).
					Return([]*model.Ad{ad}, nil)
			},
			expectErr: nil,
		},
		{
			name: "Failure - db error",
			mockBehaviour: func(a *mocks.MockAdRepository, sellerID uuid.UUID) {
				a.EXPECT().
					ListSellerAds(mock.Anything, sellerID).
					Return(nil, errors.New("db error"))
			},
			expectErr: ucerrs.ErrListSellerAdsDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adRepo := mocks.NewMockAdRepository(t)
			tt.mockBehaviour(adRepo, sellerID)

			uc := usecase.NewListSellerAdsUC(adRepo)

			_, err := uc.Execute(context.Background(), dto.ListSellerAdsInput{SellerID: sellerID})

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
