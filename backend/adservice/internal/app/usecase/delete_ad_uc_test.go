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

func TestDeleteAdUC_Execute(t *testing.T) {
	type adapter struct {
		ad        *mocks.MockAdRepository
		media     *mocks.MockMediaRepository
		publisher *mocks.MockAdPublisher
	}

	type testCase struct {
		name          string
		buildAd       func() *model.Ad // ad returned by Get
		input         func(adID uuid.UUID) dto.DeleteAdInput
		mockBehaviour func(a adapter, adID uuid.UUID)
		expectErr     error
	}

	sellerID := uuid.New()
	otherSellerID := uuid.New()

	newAd := func() *model.Ad {
		ad, _ := model.NewAd(
			sellerID, "title",
			nil, 100,
			model.CategoryFood.String(), nil,
		)
		return ad
	}
	publishedAd := func() *model.Ad {
		ad := newAd()
		_ = ad.Publish()
		return ad
	}
	// on moderation again after being published/updated once - IsOnModeration()
	// is true, but UpdatedAt != nil, so it must still go through
	// softDeleteAndPublish, not hardDelete.
	resubmittedAd := func() *model.Ad {
		ad := newAd()
		_ = ad.Publish()
		_ = ad.Update(nil, nil, nil, nil, nil)
		return ad
	}

	// Execute picks exactly one of the two scenarios (if/else):
	//  - hardDelete: ad is on moderation and was never published (UpdatedAt == nil)
	//  - softDeleteAndPublish: ad has been published at least once (or is currently published)

	var tests = []testCase{
		{
			name:    "Failure - ad not found",
			buildAd: func() *model.Ad { return nil },
			input: func(adID uuid.UUID) dto.DeleteAdInput {
				return dto.DeleteAdInput{AdID: adID, SellerID: sellerID}
			},
			mockBehaviour: func(_ adapter, _ uuid.UUID) {},
			expectErr:     ucerrs.ErrAdNotFound,
		},
		{
			name:    "Failure - db error on get",
			buildAd: func() *model.Ad { return nil },
			input: func(adID uuid.UUID) dto.DeleteAdInput {
				return dto.DeleteAdInput{AdID: adID, SellerID: sellerID}
			},
			mockBehaviour: func(_ adapter, _ uuid.UUID) {},
			expectErr:     ucerrs.ErrGetAdDB,
		},
		{
			name:    "Failure - access denied",
			buildAd: newAd,
			input: func(adID uuid.UUID) dto.DeleteAdInput {
				return dto.DeleteAdInput{AdID: adID, SellerID: otherSellerID}
			},
			mockBehaviour: func(_ adapter, _ uuid.UUID) {},
			expectErr:     ucerrs.ErrAccessDenied,
		},
		{
			name: "Failure - already deleted",
			buildAd: func() *model.Ad {
				ad := newAd()
				_ = ad.Delete()
				return ad
			},
			input: func(adID uuid.UUID) dto.DeleteAdInput {
				return dto.DeleteAdInput{AdID: adID, SellerID: sellerID}
			},
			mockBehaviour: func(_ adapter, _ uuid.UUID) {},
			expectErr:     ucerrs.ErrCannotDelete,
		},
		{
			// hardDelete's own domain ad.Delete() succeeds (first call),
			// but the DB delete inside hardDelete's transaction fails.
			name:    "Failure - delete ad db error (on moderation)",
			buildAd: newAd,
			input: func(adID uuid.UUID) dto.DeleteAdInput {
				return dto.DeleteAdInput{AdID: adID, SellerID: sellerID}
			},
			mockBehaviour: func(a adapter, adID uuid.UUID) {
				a.ad.EXPECT().Delete(mock.Anything, adID).Return(errors.New("db error"))
			},
			expectErr: ucerrs.ErrDeleteAdDB,
		},
		{
			name:    "Failure - delete images db error (on moderation)",
			buildAd: newAd,
			input: func(adID uuid.UUID) dto.DeleteAdInput {
				return dto.DeleteAdInput{AdID: adID, SellerID: sellerID}
			},
			mockBehaviour: func(a adapter, adID uuid.UUID) {
				a.ad.EXPECT().Delete(mock.Anything, adID).Return(nil)
				a.media.EXPECT().Delete(mock.Anything, adID).Return(errors.New("db error"))
			},
			expectErr: ucerrs.ErrDeleteImagesDB,
		},
		{
			// On-moderation ad, never published: only hardDelete runs,
			// softDeleteAndPublish (and its ad.Delete() call) must NOT be hit.
			name:    "Success - on moderation",
			buildAd: newAd,
			input: func(adID uuid.UUID) dto.DeleteAdInput {
				return dto.DeleteAdInput{AdID: adID, SellerID: sellerID}
			},
			mockBehaviour: func(a adapter, adID uuid.UUID) {
				a.ad.EXPECT().Delete(mock.Anything, adID).Return(nil)
				a.media.EXPECT().Delete(mock.Anything, adID).Return(nil)
			},
			expectErr: nil,
		},
		{
			// Published ad: hardDelete is skipped (not on moderation, or
			// UpdatedAt is set), so only softDeleteAndPublish runs.
			name:    "Success - published",
			buildAd: publishedAd,
			input: func(adID uuid.UUID) dto.DeleteAdInput {
				return dto.DeleteAdInput{AdID: adID, SellerID: sellerID}
			},
			mockBehaviour: func(a adapter, adID uuid.UUID) {
				a.ad.EXPECT().Update(mock.Anything, mock.AnythingOfType("*model.Ad")).Return(nil)
				a.media.EXPECT().Delete(mock.Anything, adID).Return(nil)
				a.publisher.EXPECT().PublishAdDeleted(mock.Anything, adID).Return(nil)
			},
			expectErr: nil,
		},
		{
			name:    "Failure - update ad db error (published)",
			buildAd: publishedAd,
			input: func(adID uuid.UUID) dto.DeleteAdInput {
				return dto.DeleteAdInput{AdID: adID, SellerID: sellerID}
			},
			mockBehaviour: func(a adapter, _ uuid.UUID) {
				a.ad.EXPECT().Update(mock.Anything, mock.AnythingOfType("*model.Ad")).Return(errors.New("db error"))
			},
			expectErr: ucerrs.ErrUpdateAdDB,
		},
		{
			name:    "Failure - delete images db error (published)",
			buildAd: publishedAd,
			input: func(adID uuid.UUID) dto.DeleteAdInput {
				return dto.DeleteAdInput{AdID: adID, SellerID: sellerID}
			},
			mockBehaviour: func(a adapter, adID uuid.UUID) {
				a.ad.EXPECT().Update(mock.Anything, mock.AnythingOfType("*model.Ad")).Return(nil)
				a.media.EXPECT().Delete(mock.Anything, adID).Return(errors.New("db error"))
			},
			expectErr: ucerrs.ErrDeleteImagesDB,
		},
		{
			// On-moderation ad that was previously published/updated
			// (UpdatedAt != nil): must go through softDeleteAndPublish,
			// hardDelete's uc.ad.Delete/media.Delete-only-no-publish path
			// must NOT be hit.
			name:    "Success - on moderation after update (resubmitted)",
			buildAd: resubmittedAd,
			input: func(adID uuid.UUID) dto.DeleteAdInput {
				return dto.DeleteAdInput{AdID: adID, SellerID: sellerID}
			},
			mockBehaviour: func(a adapter, adID uuid.UUID) {
				a.ad.EXPECT().Update(mock.Anything, mock.AnythingOfType("*model.Ad")).Return(nil)
				a.media.EXPECT().Delete(mock.Anything, adID).Return(nil)
				a.publisher.EXPECT().PublishAdDeleted(mock.Anything, adID).Return(nil)
			},
			expectErr: nil,
		},
		{
			name:    "Failure - publish ad deleted error (published)",
			buildAd: publishedAd,
			input: func(adID uuid.UUID) dto.DeleteAdInput {
				return dto.DeleteAdInput{AdID: adID, SellerID: sellerID}
			},
			mockBehaviour: func(a adapter, adID uuid.UUID) {
				a.ad.EXPECT().Update(mock.Anything, mock.AnythingOfType("*model.Ad")).Return(nil)
				a.media.EXPECT().Delete(mock.Anything, adID).Return(nil)
				a.publisher.EXPECT().PublishAdDeleted(mock.Anything, adID).Return(errors.New("mq error"))
			},
			expectErr: ucerrs.ErrPublishAdDeleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adRepo := mocks.NewMockAdRepository(t)
			mediaRepo := mocks.NewMockMediaRepository(t)
			publisher := mocks.NewMockAdPublisher(t)

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

			tt.mockBehaviour(adapter{ad: adRepo, media: mediaRepo, publisher: publisher}, adID)

			uc := usecase.NewDeleteAdUC(mocks.FakeTxManager{}, adRepo, mediaRepo, publisher)

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
