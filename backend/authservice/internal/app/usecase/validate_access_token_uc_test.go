package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/maket12/ads-service/authservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/authservice/internal/app/errs"
	"github.com/maket12/ads-service/authservice/internal/app/usecase"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	"github.com/maket12/ads-service/authservice/internal/domain/port/mocks"
	pkgerrs "github.com/maket12/ads-service/authservice/pkg/errs"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestValidateAccessTokenUC_Execute(t *testing.T) {
	type adapter struct {
		account        *mocks.MockAccountRepository
		tokenGenerator *mocks.MockTokenGenerator
	}

	type testCase struct {
		name          string
		input         dto.ValidateAccessTokenInput
		mockBehaviour func(a adapter, accountID uuid.UUID, account *model.Account)
		expectErr     error
	}

	tokenStr := "access-jwt"
	roleName := "user"

	var tests = []testCase{
		{
			name: "Success",
			input: dto.ValidateAccessTokenInput{
				AccessToken: tokenStr,
			},
			mockBehaviour: func(a adapter, accountID uuid.UUID, account *model.Account) {
				a.tokenGenerator.EXPECT().ValidateAccessToken(mock.Anything, tokenStr).Return(accountID, roleName, nil)
				a.account.EXPECT().GetByID(mock.Anything, accountID).Return(account, nil)
			},
			expectErr: nil,
		},
		{
			name: "Failure - parsing failed",
			input: dto.ValidateAccessTokenInput{
				AccessToken: tokenStr,
			},
			mockBehaviour: func(a adapter, accountID uuid.UUID, account *model.Account) {
				a.tokenGenerator.EXPECT().ValidateAccessToken(mock.Anything, tokenStr).Return(uuid.Nil, "", errors.New("malformed"))
			},
			expectErr: ucerrs.ErrInvalidAccessToken,
		},
		{
			name: "Failure - account records missing",
			input: dto.ValidateAccessTokenInput{
				AccessToken: tokenStr,
			},
			mockBehaviour: func(a adapter, accountID uuid.UUID, account *model.Account) {
				a.tokenGenerator.EXPECT().ValidateAccessToken(mock.Anything, tokenStr).Return(accountID, roleName, nil)
				a.account.EXPECT().GetByID(mock.Anything, accountID).Return(nil, pkgerrs.ErrObjectNotFound)
			},
			expectErr: ucerrs.ErrInvalidAccessToken,
		},
		{
			name: "Failure - database connection reset",
			input: dto.ValidateAccessTokenInput{
				AccessToken: tokenStr,
			},
			mockBehaviour: func(a adapter, accountID uuid.UUID, account *model.Account) {
				a.tokenGenerator.EXPECT().ValidateAccessToken(mock.Anything, tokenStr).Return(accountID, roleName, nil)
				a.account.EXPECT().GetByID(mock.Anything, accountID).Return(nil, errors.New("reset"))
			},
			expectErr: ucerrs.ErrGetAccountByIDDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountID := uuid.New()
			account, err := model.NewAccount("validate@example.com", "hash")
			assert.NoError(t, err)

			accountRepo := mocks.NewMockAccountRepository(t)
			tokenGenerator := mocks.NewMockTokenGenerator(t)

			tt.mockBehaviour(adapter{
				account:        accountRepo,
				tokenGenerator: tokenGenerator,
			}, accountID, account)

			uc := usecase.NewValidateAccessTokenUC(accountRepo, tokenGenerator)

			out, err := uc.Execute(context.Background(), tt.input)

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, accountID, out.AccountID)
				assert.Equal(t, roleName, out.Role)
			}
		})
	}
}
