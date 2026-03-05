package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/maket12/ads-service/authservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/authservice/internal/app/errs"
	"github.com/maket12/ads-service/authservice/internal/app/usecase"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	"github.com/maket12/ads-service/authservice/internal/domain/port/mocks"
	pkgerrs "github.com/maket12/ads-service/pkg/errs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLoginUC_Execute(t *testing.T) {
	type adapter struct {
		account        *mocks.AccountRepository
		accountRole    *mocks.AccountRoleRepository
		refreshSession *mocks.RefreshSessionRepository
		passwordHasher *mocks.PasswordHasher
		tokenGenerator *mocks.TokenGenerator
	}

	type testCase struct {
		name    string
		input   dto.LoginInput
		prepare func(a adapter)
		wantErr error
	}

	email := "user@test.com"
	pass := "password123"
	ttl := time.Hour * 24

	account, _ := model.NewAccount(email, "hashed_db")

	role, _ := model.NewAccountRole(account.ID())

	var tests = []testCase{
		{
			name:  "Success",
			input: dto.LoginInput{Email: email, Password: pass, IP: nil, UserAgent: nil},
			prepare: func(a adapter) {
				a.account.On("GetByEmail", mock.Anything, email).Return(account, nil)
				a.passwordHasher.On("Compare", "hashed_db", pass).Return(true)
				a.account.On("MarkLogin", mock.Anything, mock.Anything).Return(nil)
				a.accountRole.On("Get", mock.Anything, account.ID()).Return(role, nil)
				a.tokenGenerator.On("GenerateAccessToken", mock.Anything, account.ID(), "user").
					Return("access_token_val", nil)
				a.tokenGenerator.On("GenerateRefreshToken", mock.Anything, account.ID(), mock.Anything).
					Return("refresh_token_val", nil)

				a.refreshSession.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: nil,
		},
		{
			name:  "Fail - account Not Found",
			input: dto.LoginInput{Email: "unknown@test.com", Password: pass},
			prepare: func(a adapter) {
				a.account.On("GetByEmail", mock.Anything, "unknown@test.com").
					Return(nil, pkgerrs.ErrObjectNotFound)
			},
			wantErr: ucerrs.ErrInvalidCredentials,
		},
		{
			name:  "Fail - Password Mismatch",
			input: dto.LoginInput{Email: email, Password: "wrong_password"},
			prepare: func(a adapter) {
				a.account.On("GetByEmail", mock.Anything, email).Return(account, nil)
				a.passwordHasher.On("Compare", "hashed_db", "wrong_password").Return(false)
			},
			wantErr: ucerrs.ErrInvalidCredentials,
		},
		{
			name:  "Fail - account Banned",
			input: dto.LoginInput{Email: email, Password: pass},
			prepare: func(a adapter) {
				bannedAcc, _ := model.NewAccount(email, "hashed_db")
				bannedAcc.Block()

				a.account.On("GetByEmail", mock.Anything, email).Return(bannedAcc, nil)
				a.passwordHasher.On("Compare", "hashed_db", pass).Return(true)
			},
			wantErr: ucerrs.ErrCannotLogin,
		},
		{
			name:  "Fail - Token Generation Error",
			input: dto.LoginInput{Email: email, Password: pass},
			prepare: func(a adapter) {
				a.account.On("GetByEmail", mock.Anything, email).Return(account, nil)
				a.passwordHasher.On("Compare", "hashed_db", pass).Return(true)
				a.account.On("MarkLogin", mock.Anything, mock.Anything).Return(nil)
				a.accountRole.On("Get", mock.Anything, account.ID()).Return(role, nil)

				a.tokenGenerator.On("GenerateAccessToken", mock.Anything, mock.Anything, mock.Anything).
					Return("", assert.AnError)
			},
			wantErr: ucerrs.ErrGenerateAccessToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := adapter{
				account:        mocks.NewAccountRepository(t),
				accountRole:    mocks.NewAccountRoleRepository(t),
				refreshSession: mocks.NewRefreshSessionRepository(t),
				passwordHasher: mocks.NewPasswordHasher(t),
				tokenGenerator: mocks.NewTokenGenerator(t),
			}

			tt.prepare(a)

			uc := usecase.NewLoginUC(
				a.account, a.accountRole, a.refreshSession,
				a.passwordHasher, a.tokenGenerator, ttl,
			)

			res, err := uc.Execute(context.Background(), tt.input)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Empty(t, res.AccessToken)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "access_token_val", res.AccessToken)
				assert.Equal(t, "refresh_token_val", res.RefreshToken)
			}
		})
	}
}
