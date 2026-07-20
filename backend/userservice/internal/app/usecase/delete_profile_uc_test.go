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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeleteProfileUC_Execute(t *testing.T) {
	type adapter struct {
		profile *mocks.MockProfileRepository
	}

	type testCase struct {
		name          string
		mockBehaviour func(a adapter)
		expectErr     error
	}

	accountID := uuid.New()

	var tests = []testCase{
		{
			name: "Success",
			mockBehaviour: func(a adapter) {
				a.profile.EXPECT().
					Delete(mock.Anything, accountID).
					Return(nil)
			},
			expectErr: nil,
		},
		{
			name: "Failure - db error on Delete",
			mockBehaviour: func(a adapter) {
				a.profile.EXPECT().
					Delete(mock.Anything, accountID).
					Return(errors.New("db error"))
			},
			expectErr: ucerrs.ErrDeleteProfileDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profileRepo := mocks.NewMockProfileRepository(t)

			tt.mockBehaviour(adapter{profile: profileRepo})

			uc := usecase.NewDeleteProfileUC(profileRepo)

			out, err := uc.Execute(context.Background(), dto.DeleteProfileInput{AccountID: accountID})

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
				assert.True(t, out.Deleted)
			}
		})
	}
}
