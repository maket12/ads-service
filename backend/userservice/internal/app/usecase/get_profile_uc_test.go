package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/maket12/ads-service/userservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/userservice/internal/app/errs"
	"github.com/maket12/ads-service/userservice/internal/app/usecase"
	"github.com/maket12/ads-service/userservice/internal/domain/model"
	"github.com/maket12/ads-service/userservice/internal/domain/port/mocks"
	pkgerrs "github.com/maket12/ads-service/userservice/pkg/errs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetProfileUC_Execute(t *testing.T) {
	type adapter struct {
		profile *mocks.MockProfileRepository
	}

	type testCase struct {
		name          string
		mockBehaviour func(a adapter, p *model.Profile)
		expectErr     error
	}

	accountID := uuid.New()

	var tests = []testCase{
		{
			name: "Success",
			mockBehaviour: func(a adapter, p *model.Profile) {
				a.profile.EXPECT().
					Get(mock.Anything, accountID).
					Return(p, nil)
			},
			expectErr: nil,
		},
		{
			name: "Failure - profile not found",
			mockBehaviour: func(a adapter, p *model.Profile) {
				a.profile.EXPECT().
					Get(mock.Anything, accountID).
					Return(nil, pkgerrs.ErrObjectNotFound)
			},
			expectErr: ucerrs.ErrProfileNotFound,
		},
		{
			name: "Failure - db error on Get",
			mockBehaviour: func(a adapter, p *model.Profile) {
				a.profile.EXPECT().
					Get(mock.Anything, accountID).
					Return(nil, errors.New("db error"))
			},
			expectErr: ucerrs.ErrGetProfileDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile, err := model.NewProfile(accountID)
			assert.NoError(t, err)

			profileRepo := mocks.NewMockProfileRepository(t)

			tt.mockBehaviour(adapter{profile: profileRepo}, profile)

			uc := usecase.NewGetProfileUC(profileRepo)

			out, err := uc.Execute(context.Background(), dto.GetProfileInput{AccountID: accountID})

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, accountID, out.AccountID)
			}
		})
	}
}
