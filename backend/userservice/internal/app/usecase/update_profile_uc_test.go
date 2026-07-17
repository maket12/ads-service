package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
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

func TestUpdateProfileUC_Execute(t *testing.T) {
	type adapter struct {
		profile *mocks.MockProfileRepository
		phone   *mocks.MockPhoneValidator
	}

	type testCase struct {
		name          string
		input         dto.UpdateProfileInput
		mockBehaviour func(a adapter, p *model.Profile)
		expectErr     error
	}

	accountID := uuid.New()
	firstName := gofakeit.FirstName()
	lastName := gofakeit.LastName()
	rawPhone := gofakeit.Phone()
	normalizedPhone := "+1" + rawPhone
	avatarURL := gofakeit.URL()
	bio := gofakeit.Bio()

	var tests = []testCase{
		{
			name: "Success - without phone",
			input: dto.UpdateProfileInput{
				AccountID: accountID,
				FirstName: &firstName,
				LastName:  &lastName,
				AvatarURL: &avatarURL,
				Bio:       &bio,
			},
			mockBehaviour: func(a adapter, p *model.Profile) {
				a.profile.EXPECT().
					Get(mock.Anything, accountID).
					Return(p, nil)

				a.profile.EXPECT().
					Update(mock.Anything, p).
					Return(nil)
			},
			expectErr: nil,
		},
		{
			name: "Success - with phone",
			input: dto.UpdateProfileInput{
				AccountID: accountID,
				FirstName: &firstName,
				LastName:  &lastName,
				Phone:     &rawPhone,
				AvatarURL: &avatarURL,
				Bio:       &bio,
			},
			mockBehaviour: func(a adapter, p *model.Profile) {
				a.profile.EXPECT().
					Get(mock.Anything, accountID).
					Return(p, nil)

				a.phone.EXPECT().
					Validate(mock.Anything, rawPhone).
					Return(normalizedPhone, nil)

				a.profile.EXPECT().
					Update(mock.Anything, p).
					Return(nil)
			},
			expectErr: nil,
		},
		{
			name: "Failure - profile not found",
			input: dto.UpdateProfileInput{
				AccountID: accountID,
			},
			mockBehaviour: func(a adapter, p *model.Profile) {
				a.profile.EXPECT().
					Get(mock.Anything, accountID).
					Return(nil, pkgerrs.ErrObjectNotFound)
			},
			expectErr: ucerrs.ErrProfileNotFound,
		},
		{
			name: "Failure - db error on Get",
			input: dto.UpdateProfileInput{
				AccountID: accountID,
			},
			mockBehaviour: func(a adapter, p *model.Profile) {
				a.profile.EXPECT().
					Get(mock.Anything, accountID).
					Return(nil, errors.New("db error"))
			},
			expectErr: ucerrs.ErrGetProfileDB,
		},
		{
			name: "Failure - invalid phone",
			input: dto.UpdateProfileInput{
				AccountID: accountID,
				Phone:     &rawPhone,
			},
			mockBehaviour: func(a adapter, p *model.Profile) {
				a.profile.EXPECT().
					Get(mock.Anything, accountID).
					Return(p, nil)

				a.phone.EXPECT().
					Validate(mock.Anything, rawPhone).
					Return("", errors.New("invalid phone format"))
			},
			expectErr: ucerrs.ErrInvalidInput,
		},
		{
			name: "Failure - db error on Update",
			input: dto.UpdateProfileInput{
				AccountID: accountID,
				FirstName: &firstName,
				LastName:  &lastName,
				AvatarURL: &avatarURL,
				Bio:       &bio,
			},
			mockBehaviour: func(a adapter, p *model.Profile) {
				a.profile.EXPECT().
					Get(mock.Anything, accountID).
					Return(p, nil)

				a.profile.EXPECT().
					Update(mock.Anything, p).
					Return(errors.New("db error"))
			},
			expectErr: ucerrs.ErrUpdateProfileDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile, err := model.NewProfile(accountID)
			assert.NoError(t, err)

			profileRepo := mocks.NewMockProfileRepository(t)
			phoneValidator := mocks.NewMockPhoneValidator(t)

			tt.mockBehaviour(adapter{profile: profileRepo, phone: phoneValidator}, profile)

			uc := usecase.NewUpdateProfileUC(profileRepo, phoneValidator)

			out, err := uc.Execute(context.Background(), tt.input)

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
				assert.True(t, out.Success)
			}
		})
	}
}
