package usecase_test

import (
	"context"
	"github.com/maket12/ads-service/authservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/authservice/internal/app/errs"
	"github.com/maket12/ads-service/authservice/internal/app/usecase"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	"github.com/maket12/ads-service/authservice/internal/domain/port/mocks"
	pkgerrs "github.com/maket12/ads-service/pkg/errs"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestValidateAccessTokenUC_Execute(t *testing.T) {
	type adapter struct {
		account        *mocks.AccountRepository
		tokenGenerator *mocks.TokenGenerator
	}

	type testCase struct {
		name    string
		input   dto.ValidateAccessTokenInput
		prepare func(a adapter)
		wantErr error
	}

	accountID := uuid.New()
	role := "user"
	accessToken := "valid-access-token"

	activeAcc, _ := model.NewAccount("test@test.com", "hash")

	bannedAcc, _ := model.NewAccount("banned@test.com", "hash")
	bannedAcc.Block()

	var tests = []testCase{
		{
			name: "Success",
			input: dto.ValidateAccessTokenInput{
				AccessToken: accessToken,
			},
			prepare: func(a adapter) {
				a.tokenGenerator.On("ValidateAccessTokenInput", mock.Anything, accessToken).
					Return(accountID, role, nil)
				a.account.On("GetByID", mock.Anything, accountID).
					Return(activeAcc, nil)
			},
			wantErr: nil,
		},
		{
			name: "Fail - Invalid Token",
			input: dto.ValidateAccessTokenInput{
				AccessToken: "expired-or-fake-token",
			},
			prepare: func(a adapter) {
				a.tokenGenerator.On("ValidateAccessTokenInput", mock.Anything, "expired-or-fake-token").
					Return(uuid.Nil, "", assert.AnError)
			},
			wantErr: ucerrs.ErrInvalidAccessToken,
		},
		{
			name: "Fail - account Not Found In DB",
			input: dto.ValidateAccessTokenInput{
				AccessToken: accessToken,
			},
			prepare: func(a adapter) {
				a.tokenGenerator.On("ValidateAccessTokenInput", mock.Anything, accessToken).
					Return(accountID, role, nil)
				a.account.On("GetByID", mock.Anything, accountID).
					Return(nil, pkgerrs.ErrObjectNotFound)
			},
			wantErr: ucerrs.ErrInvalidAccessToken,
		},
		{
			name: "Fail - account Is Banned",
			input: dto.ValidateAccessTokenInput{
				AccessToken: accessToken,
			},
			prepare: func(a adapter) {
				a.tokenGenerator.On("ValidateAccessTokenInput", mock.Anything, accessToken).
					Return(accountID, role, nil)
				a.account.On("GetByID", mock.Anything, accountID).
					Return(bannedAcc, nil)
			},
			wantErr: ucerrs.ErrCannotLogin,
		},
		{
			name: "Fail - Database Error",
			input: dto.ValidateAccessTokenInput{
				AccessToken: accessToken,
			},
			prepare: func(a adapter) {
				a.tokenGenerator.On("ValidateAccessTokenInput", mock.Anything, accessToken).
					Return(accountID, role, nil)
				a.account.On("GetByID", mock.Anything, accountID).
					Return(nil, assert.AnError)
			},
			wantErr: ucerrs.ErrGetAccountByIDDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := adapter{
				account:        mocks.NewAccountRepository(t),
				tokenGenerator: mocks.NewTokenGenerator(t),
			}

			tt.prepare(a)

			uc := usecase.NewValidateAccessTokenUC(a.account, a.tokenGenerator)

			res, err := uc.Execute(context.Background(), tt.input)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Empty(t, res.Role)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, role, res.Role)
				assert.Equal(t, accountID, res.AccountID)
			}
		})
	}
}
