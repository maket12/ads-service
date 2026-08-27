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

func TestDeleteAllAdsUC_Execute(t *testing.T) {
	type adapter struct {
		ad    *mocks.MockAdRepository
		media *mocks.MockMediaRepository
	}

	type testCase struct {
		name          string
		buildAds      func() []*model.Ad
		mockBehaviour func(a adapter, ads []*model.Ad)
		expectErr     error
	}

	sellerID := uuid.New()

	moderationAd := func() *model.Ad {
		ad, _ := model.NewAd(
			sellerID, "title",
			nil, 100,
			model.CategoryFood.String(), nil,
		)
		return ad
	}
	publishedAd := func() *model.Ad {
		ad := moderationAd()
		_ = ad.Publish()
		return ad
	}

	var tests = []testCase{
		{
			name: "Success - mixed moderation and published ads",
			buildAds: func() []*model.Ad {
				return []*model.Ad{moderationAd(), publishedAd()}
			},
			mockBehaviour: func(a adapter, ads []*model.Ad) {
				a.ad.EXPECT().Delete(mock.Anything, ads[0].ID()).Return(nil)
				a.media.EXPECT().Delete(mock.Anything, ads[0].ID()).Return(nil)
				a.ad.EXPECT().Update(mock.Anything, ads[1]).Return(nil)
			},
			expectErr: nil,
		},
		{
			name: "Failure - list seller ads db error",
			buildAds: func() []*model.Ad {
				return nil
			},
			mockBehaviour: func(_ adapter, _ []*model.Ad) {},
			expectErr:     ucerrs.ErrListSellerAdsDB,
		},
		{
			name: "Failure - delete ad db error (on moderation)",
			buildAds: func() []*model.Ad {
				return []*model.Ad{moderationAd()}
			},
			mockBehaviour: func(a adapter, ads []*model.Ad) {
				a.ad.EXPECT().Delete(mock.Anything, ads[0].ID()).Return(errors.New("db error"))
			},
			expectErr: ucerrs.ErrDeleteAdDB,
		},
		{
			name: "Failure - delete images db error (on moderation)",
			buildAds: func() []*model.Ad {
				return []*model.Ad{moderationAd()}
			},
			mockBehaviour: func(a adapter, ads []*model.Ad) {
				a.ad.EXPECT().Delete(mock.Anything, ads[0].ID()).Return(nil)
				a.media.EXPECT().Delete(mock.Anything, ads[0].ID()).Return(errors.New("db error"))
			},
			expectErr: ucerrs.ErrDeleteImagesDB,
		},
		{
			name: "Failure - update ad db error (published)",
			buildAds: func() []*model.Ad {
				return []*model.Ad{publishedAd()}
			},
			mockBehaviour: func(a adapter, ads []*model.Ad) {
				a.ad.EXPECT().Update(mock.Anything, ads[0]).Return(errors.New("db error"))
			},
			expectErr: ucerrs.ErrUpdateAdDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adRepo := mocks.NewMockAdRepository(t)
			mediaRepo := mocks.NewMockMediaRepository(t)

			ads := tt.buildAds()

			if tt.name == "Failure - list seller ads db error" {
				adRepo.EXPECT().ListSellerAds(mock.Anything, sellerID).Return(nil, errors.New("db error"))
			} else {
				adRepo.EXPECT().ListSellerAds(mock.Anything, sellerID).Return(ads, nil)
			}

			tt.mockBehaviour(adapter{ad: adRepo, media: mediaRepo}, ads)

			uc := usecase.NewDeleteAllAdsUC(mocks.FakeTxManager{}, adRepo, mediaRepo)

			out, err := uc.Execute(context.Background(), dto.DeleteAllAdsInput{SellerID: sellerID})

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
				assert.False(t, out.Success)
			} else {
				assert.NoError(t, err)
				assert.True(t, out.Success)
			}
		})
	}
}
