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
	pkgerrs "github.com/maket12/ads-service/adservice/pkg/errs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAdUC_Execute(t *testing.T) {
	type adapter struct {
		ad    *mocks.MockAdRepository
		media *mocks.MockMediaRepository
	}

	type testCase struct {
		name          string
		buildAd       func() *model.Ad
		input         func(adID uuid.UUID) dto.GetAdInput
		mockBehaviour func(a adapter, adID uuid.UUID)
		expectErr     error
	}

	sellerID := uuid.New()
	otherSellerID := uuid.New()
	images := []string{"img1.png"}

	newAd := func() *model.Ad {
		ad, _ := model.NewAd(sellerID, "title", nil, 100, nil)
		return ad
	}
	publishedAd := func() *model.Ad {
		ad := newAd()
		_ = ad.Publish()
		return ad
	}

	var tests = []testCase{
		{
			name:    "Success - published ad, any viewer",
			buildAd: publishedAd,
			input: func(adID uuid.UUID) dto.GetAdInput {
				return dto.GetAdInput{AdID: adID, SellerID: otherSellerID}
			},
			mockBehaviour: func(a adapter, adID uuid.UUID) {
				a.media.EXPECT().Get(mock.Anything, adID).Return(images, nil)
			},
			expectErr: nil,
		},
		{
			name:    "Success - unpublished ad, owner",
			buildAd: newAd,
			input: func(adID uuid.UUID) dto.GetAdInput {
				return dto.GetAdInput{AdID: adID, SellerID: sellerID}
			},
			mockBehaviour: func(a adapter, adID uuid.UUID) {
				a.media.EXPECT().Get(mock.Anything, adID).Return(images, nil)
			},
			expectErr: nil,
		},
		{
			name:    "Failure - ad not found",
			buildAd: func() *model.Ad { return nil },
			input: func(adID uuid.UUID) dto.GetAdInput {
				return dto.GetAdInput{AdID: adID, SellerID: sellerID}
			},
			mockBehaviour: func(_ adapter, _ uuid.UUID) {},
			expectErr:     ucerrs.ErrAdNotFound,
		},
		{
			name:    "Failure - db error on get",
			buildAd: func() *model.Ad { return nil },
			input: func(adID uuid.UUID) dto.GetAdInput {
				return dto.GetAdInput{AdID: adID, SellerID: sellerID}
			},
			mockBehaviour: func(_ adapter, _ uuid.UUID) {},
			expectErr:     ucerrs.ErrGetAdDB,
		},
		{
			name:    "Failure - access denied (unpublished, not owner)",
			buildAd: newAd,
			input: func(adID uuid.UUID) dto.GetAdInput {
				return dto.GetAdInput{AdID: adID, SellerID: otherSellerID}
			},
			mockBehaviour: func(_ adapter, _ uuid.UUID) {},
			expectErr:     ucerrs.ErrAccessDenied,
		},
		{
			name:    "Failure - get images db error",
			buildAd: publishedAd,
			input: func(adID uuid.UUID) dto.GetAdInput {
				return dto.GetAdInput{AdID: adID, SellerID: otherSellerID}
			},
			mockBehaviour: func(a adapter, adID uuid.UUID) {
				a.media.EXPECT().Get(mock.Anything, adID).Return(nil, errors.New("db error"))
			},
			expectErr: ucerrs.ErrGetImagesDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adRepo := mocks.NewMockAdRepository(t)
			mediaRepo := mocks.NewMockMediaRepository(t)

			ad := tt.buildAd()
			adID := uuid.New()
			if ad != nil {
				adID = ad.ID()
			}

			switch tt.name {
			case "Failure - ad not found":
				adRepo.EXPECT().Get(mock.Anything, adID).Return(nil, pkgerrs.ErrObjectNotFound)
			case "Failure - db error on get":
				adRepo.EXPECT().Get(mock.Anything, adID).Return(nil, errors.New("db error"))
			default:
				adRepo.EXPECT().Get(mock.Anything, adID).Return(ad, nil)
			}

			tt.mockBehaviour(adapter{ad: adRepo, media: mediaRepo}, adID)

			uc := usecase.NewGetAdUC(adRepo, mediaRepo)

			_, err := uc.Execute(context.Background(), tt.input(adID))

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
