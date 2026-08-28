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
	pkgerrs "github.com/maket12/ads-service/backend/authservice/pkg/errs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateAdUC_Execute(t *testing.T) {
	type adapter struct {
		ad    *mocks.MockAdRepository
		media *mocks.MockMediaRepository
	}

	type testCase struct {
		name          string
		buildAd       func() *model.Ad
		input         func(adID uuid.UUID) dto.UpdateAdInput
		mockBehaviour func(a adapter, adID uuid.UUID)
		expectErr     error
	}

	sellerID := uuid.New()
	otherSellerID := uuid.New()
	newTitle := "new title"
	newDescription := "new description"
	newPrice := int64(200)
	newImages := []string{"img2.png"}

	newAd := func() *model.Ad {
		ad, _ := model.NewAd(
			sellerID, "title",
			nil, 100,
			model.CategoryFood.String(), nil,
		)
		return ad
	}
	// Ad that can no longer be updated, e.g. already deleted.
	undeployableAd := func() *model.Ad {
		ad := newAd()
		_ = ad.Delete()
		return ad
	}

	var tests = []testCase{
		{
			name:    "Success",
			buildAd: newAd,
			input: func(adID uuid.UUID) dto.UpdateAdInput {
				return dto.UpdateAdInput{
					AdID:        adID,
					SellerID:    sellerID,
					Title:       &newTitle,
					Description: &newDescription,
					Price:       &newPrice,
					Images:      newImages,
				}
			},
			mockBehaviour: func(a adapter, adID uuid.UUID) {
				a.ad.EXPECT().Update(mock.Anything, mock.AnythingOfType("*model.Ad")).Return(nil)
				a.media.EXPECT().Save(mock.Anything, adID, newImages).Return(nil)
			},
			expectErr: nil,
		},
		{
			name:    "Failure - ad not found",
			buildAd: func() *model.Ad { return nil },
			input: func(adID uuid.UUID) dto.UpdateAdInput {
				return dto.UpdateAdInput{AdID: adID, SellerID: sellerID}
			},
			mockBehaviour: func(_ adapter, _ uuid.UUID) {},
			expectErr:     ucerrs.ErrAdNotFound,
		},
		{
			name:    "Failure - db error on get",
			buildAd: func() *model.Ad { return nil },
			input: func(adID uuid.UUID) dto.UpdateAdInput {
				return dto.UpdateAdInput{AdID: adID, SellerID: sellerID}
			},
			mockBehaviour: func(_ adapter, _ uuid.UUID) {},
			expectErr:     ucerrs.ErrGetAdDB,
		},
		{
			name:    "Failure - access denied",
			buildAd: newAd,
			input: func(adID uuid.UUID) dto.UpdateAdInput {
				return dto.UpdateAdInput{AdID: adID, SellerID: otherSellerID}
			},
			mockBehaviour: func(_ adapter, _ uuid.UUID) {},
			expectErr:     ucerrs.ErrAccessDenied,
		},
		{
			name:    "Failure - cannot update (ad deleted)",
			buildAd: undeployableAd,
			input: func(adID uuid.UUID) dto.UpdateAdInput {
				return dto.UpdateAdInput{
					AdID:        adID,
					SellerID:    sellerID,
					Title:       &newTitle,
					Description: &newDescription,
					Price:       &newPrice,
					Images:      newImages,
				}
			},
			mockBehaviour: func(_ adapter, _ uuid.UUID) {},
			expectErr:     ucerrs.ErrCannotUpdate,
		},
		{
			name:    "Failure - update ad db error",
			buildAd: newAd,
			input: func(adID uuid.UUID) dto.UpdateAdInput {
				return dto.UpdateAdInput{
					AdID:        adID,
					SellerID:    sellerID,
					Title:       &newTitle,
					Description: &newDescription,
					Price:       &newPrice,
					Images:      newImages,
				}
			},
			mockBehaviour: func(a adapter, _ uuid.UUID) {
				a.ad.EXPECT().Update(mock.Anything, mock.AnythingOfType("*model.Ad")).Return(errors.New("db error"))
			},
			expectErr: ucerrs.ErrUpdateAdDB,
		},
		{
			name:    "Failure - save images db error",
			buildAd: newAd,
			input: func(adID uuid.UUID) dto.UpdateAdInput {
				return dto.UpdateAdInput{
					AdID:        adID,
					SellerID:    sellerID,
					Title:       &newTitle,
					Description: &newDescription,
					Price:       &newPrice,
					Images:      newImages,
				}
			},
			mockBehaviour: func(a adapter, adID uuid.UUID) {
				a.ad.EXPECT().Update(mock.Anything, mock.AnythingOfType("*model.Ad")).Return(nil)
				a.media.EXPECT().Save(mock.Anything, adID, newImages).Return(errors.New("db error"))
			},
			expectErr: ucerrs.ErrSaveImagesDB,
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

			uc := usecase.NewUpdateAdUC(mocks.FakeTxManager{}, adRepo, mediaRepo)

			out, err := uc.Execute(context.Background(), tt.input(adID))

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
