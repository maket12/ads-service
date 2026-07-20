package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/maket12/ads-service/userservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/userservice/internal/app/errs"
	"github.com/maket12/ads-service/userservice/internal/app/usecase"
	"github.com/maket12/ads-service/userservice/internal/domain/port/mocks"
	pkgerrs "github.com/maket12/ads-service/userservice/pkg/errs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateProfileUC_Execute(t *testing.T) {
	type adapter struct {
		profile *mocks.MockProfileRepository
	}

	type testCase struct {
		name          string
		input         dto.CreateProfileInput
		mockBehaviour func(a adapter)
		expectErr     error
	}

	accountID := uuid.New()

	var tests = []testCase{
		{
			name: "Success",
			input: dto.CreateProfileInput{
				AccountID: accountID,
			},
			mockBehaviour: func(a adapter) {
				a.profile.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*model.Profile")).
					Return(nil)
			},
			expectErr: nil,
		},
		{
			name: "Success - profile already exists is treated as no-op",
			input: dto.CreateProfileInput{
				AccountID: accountID,
			},
			mockBehaviour: func(a adapter) {
				a.profile.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*model.Profile")).
					Return(pkgerrs.ErrObjectAlreadyExists)
			},
			expectErr: nil,
		},
		{
			name: "Failure - db error on Create",
			input: dto.CreateProfileInput{
				AccountID: accountID,
			},
			mockBehaviour: func(a adapter) {
				a.profile.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*model.Profile")).
					Return(errors.New("db error"))
			},
			expectErr: ucerrs.ErrCreateProfileDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileRepo := mocks.NewMockProfileRepository(t)

			tt.mockBehaviour(adapter{profile: profileRepo})

			uc := usecase.NewCreateProfileUC(profileRepo)

			err := uc.Execute(context.Background(), tt.input)

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
