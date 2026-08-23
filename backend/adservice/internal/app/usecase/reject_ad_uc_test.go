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

func TestRejectAdUC_Execute(t *testing.T) {
	type testCase struct {
		name          string
		buildAd       func() *model.Ad
		input         func(adID uuid.UUID) dto.RejectAdInput
		mockBehaviour func(a *mocks.MockAdRepository, adID uuid.UUID)
		expectErr     error
	}

	sellerID := uuid.New()

	newAd := func() *model.Ad {
		ad, _ := model.NewAd(sellerID, "title", nil, 100, nil)
		return ad
	}
	// Ad that can no longer be rejected, e.g. already deleted.
	unrejectableAd := func() *model.Ad {
		ad := newAd()
		_ = ad.Delete()
		return ad
	}

	var tests = []testCase{
		{
			name:    "Success",
			buildAd: newAd,
			input: func(adID uuid.UUID) dto.RejectAdInput {
				return dto.RejectAdInput{AdID: adID}
			},
			mockBehaviour: func(a *mocks.MockAdRepository, _ uuid.UUID) {
				a.EXPECT().Update(mock.Anything, mock.AnythingOfType("*model.Ad")).Return(nil)
			},
			expectErr: nil,
		},
		{
			name:    "Failure - ad not found",
			buildAd: func() *model.Ad { return nil },
			input: func(adID uuid.UUID) dto.RejectAdInput {
				return dto.RejectAdInput{AdID: adID}
			},
			mockBehaviour: func(_ *mocks.MockAdRepository, _ uuid.UUID) {},
			expectErr:     ucerrs.ErrAdNotFound,
		},
		{
			name:    "Failure - db error on get",
			buildAd: func() *model.Ad { return nil },
			input: func(adID uuid.UUID) dto.RejectAdInput {
				return dto.RejectAdInput{AdID: adID}
			},
			mockBehaviour: func(_ *mocks.MockAdRepository, _ uuid.UUID) {},
			expectErr:     ucerrs.ErrGetAdDB,
		},
		{
			name:    "Failure - cannot reject",
			buildAd: unrejectableAd,
			input: func(adID uuid.UUID) dto.RejectAdInput {
				return dto.RejectAdInput{AdID: adID}
			},
			mockBehaviour: func(_ *mocks.MockAdRepository, _ uuid.UUID) {},
			expectErr:     ucerrs.ErrCannotReject,
		},
		{
			name:    "Failure - update ad db error",
			buildAd: newAd,
			input: func(adID uuid.UUID) dto.RejectAdInput {
				return dto.RejectAdInput{AdID: adID}
			},
			mockBehaviour: func(a *mocks.MockAdRepository, _ uuid.UUID) {
				a.EXPECT().Update(mock.Anything, mock.AnythingOfType("*model.Ad")).Return(errors.New("db error"))
			},
			expectErr: ucerrs.ErrUpdateAdDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adRepo := mocks.NewMockAdRepository(t)

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

			tt.mockBehaviour(adRepo, adID)

			uc := usecase.NewRejectAdUC(adRepo)

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
