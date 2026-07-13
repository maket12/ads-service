package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/maket12/ads-service/authservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/authservice/internal/app/errs"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	"github.com/maket12/ads-service/authservice/internal/domain/port"
	"github.com/maket12/ads-service/authservice/internal/domain/port/mocks"
	pkgerrs "github.com/maket12/ads-service/authservice/pkg/errs"
	"github.com/maket12/ads-service/authservice/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLoginUC_Execute(t *testing.T) {
	type adapter struct {
		account        *mocks.MockAccountRepository
		accountRole    *mocks.MockAccountRoleRepository
		refreshSession *mocks.MockRefreshSessionRepository
		passwordHasher *mocks.MockPasswordHasher
		tokenGenerator *mocks.MockTokenGenerator
	}

	type testCase struct {
		name          string
		input         dto.LoginInput
		mockBehaviour func(a adapter, acc *model.Account)
		expectErr     error
	}

	email := "user@example.com"
	password := "correct-password"
	hashedPassword := "hashed-password"
	ttl := 24 * time.Hour

	var tests = []testCase{
		{
			name: "Success",
			input: dto.LoginInput{
				Email:     email,
				Password:  password,
				IP:        utils.VPtr("1.2.3.4"),
				UserAgent: utils.VPtr("Mozilla"),
			},
			mockBehaviour: func(a adapter, acc *model.Account) {
				a.account.EXPECT().
					GetByEmail(mock.Anything, email).
					Return(acc, nil)

				a.passwordHasher.EXPECT().
					Compare(hashedPassword, password).
					Return(true)

				a.account.EXPECT().
					Update(mock.Anything, acc).
					Return(nil)

				a.accountRole.EXPECT().
					Get(mock.Anything, acc.ID()).
					Return(model.RestoreAccountRole(acc.ID(), model.RoleUser), nil)

				a.tokenGenerator.EXPECT().
					GeneratePair(mock.Anything, acc.ID(), model.RoleUser.String(), mock.AnythingOfType("uuid.UUID")).
					Return(&port.TokensPair{Access: "access", Refresh: "refresh"}, nil)

				a.refreshSession.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*model.RefreshSession")).
					Return(nil)
			},
			expectErr: nil,
		},
		{
			name: "Failure - account not found",
			input: dto.LoginInput{
				Email:    email,
				Password: password,
			},
			mockBehaviour: func(a adapter, acc *model.Account) {
				a.account.EXPECT().
					GetByEmail(mock.Anything, email).
					Return(nil, pkgerrs.ErrObjectNotFound)
			},
			expectErr: ucerrs.ErrInvalidCredentials,
		},
		{
			name: "Failure - db error on GetByEmail",
			input: dto.LoginInput{
				Email:    email,
				Password: password,
			},
			mockBehaviour: func(a adapter, acc *model.Account) {
				a.account.EXPECT().
					GetByEmail(mock.Anything, email).
					Return(nil, errors.New("db error"))
			},
			expectErr: ucerrs.ErrGetAccountByEmailDB,
		},
		{
			name: "Failure - password mismatch",
			input: dto.LoginInput{
				Email:    email,
				Password: "wrong-password",
			},
			mockBehaviour: func(a adapter, acc *model.Account) {
				a.account.EXPECT().
					GetByEmail(mock.Anything, email).
					Return(acc, nil)

				a.passwordHasher.EXPECT().
					Compare(hashedPassword, "wrong-password").
					Return(false)
			},
			expectErr: ucerrs.ErrInvalidCredentials,
		},
		{
			name: "Failure - account update db error",
			input: dto.LoginInput{
				Email:    email,
				Password: password,
			},
			mockBehaviour: func(a adapter, acc *model.Account) {
				a.account.EXPECT().
					GetByEmail(mock.Anything, email).
					Return(acc, nil)

				a.passwordHasher.EXPECT().
					Compare(hashedPassword, password).
					Return(true)

				a.account.EXPECT().
					Update(mock.Anything, acc).
					Return(errors.New("db error"))
			},
			expectErr: ucerrs.ErrUpdateAccountDB,
		},
		{
			name: "Failure - get account role db error",
			input: dto.LoginInput{
				Email:    email,
				Password: password,
			},
			mockBehaviour: func(a adapter, acc *model.Account) {
				a.account.EXPECT().
					GetByEmail(mock.Anything, email).
					Return(acc, nil)

				a.passwordHasher.EXPECT().
					Compare(hashedPassword, password).
					Return(true)

				a.account.EXPECT().
					Update(mock.Anything, acc).
					Return(nil)

				a.accountRole.EXPECT().
					Get(mock.Anything, acc.ID()).
					Return(nil, errors.New("db error"))
			},
			expectErr: ucerrs.ErrGetAccountRoleDB,
		},
		{
			name: "Failure - token generation error",
			input: dto.LoginInput{
				Email:    email,
				Password: password,
			},
			mockBehaviour: func(a adapter, acc *model.Account) {
				a.account.EXPECT().
					GetByEmail(mock.Anything, email).
					Return(acc, nil)

				a.passwordHasher.EXPECT().
					Compare(hashedPassword, password).
					Return(true)

				a.account.EXPECT().
					Update(mock.Anything, acc).
					Return(nil)

				a.accountRole.EXPECT().
					Get(mock.Anything, acc.ID()).
					Return(model.RestoreAccountRole(acc.ID(), model.RoleUser), nil)

				a.tokenGenerator.EXPECT().
					GeneratePair(mock.Anything, acc.ID(), model.RoleUser.String(), mock.AnythingOfType("uuid.UUID")).
					Return(nil, errors.New("crypto failure"))
			},
			expectErr: ucerrs.ErrGenerateTokensPair,
		},
		{
			name: "Failure - create refresh session db error",
			input: dto.LoginInput{
				Email:    email,
				Password: password,
			},
			mockBehaviour: func(a adapter, acc *model.Account) {
				a.account.EXPECT().
					GetByEmail(mock.Anything, email).
					Return(acc, nil)

				a.passwordHasher.EXPECT().
					Compare(hashedPassword, password).
					Return(true)

				a.account.EXPECT().
					Update(mock.Anything, acc).
					Return(nil)

				a.accountRole.EXPECT().
					Get(mock.Anything, acc.ID()).
					Return(model.RestoreAccountRole(acc.ID(), model.RoleUser), nil)

				a.tokenGenerator.EXPECT().
					GeneratePair(mock.Anything, acc.ID(), model.RoleUser.String(), mock.AnythingOfType("uuid.UUID")).
					Return(&port.TokensPair{Access: "access", Refresh: "refresh"}, nil)

				a.refreshSession.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*model.RefreshSession")).
					Return(errors.New("db error"))
			},
			expectErr: ucerrs.ErrCreateRefreshSessionDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc, err := model.NewAccount(email, hashedPassword)
			assert.NoError(t, err)

			accountRepo := mocks.NewMockAccountRepository(t)
			accountRoleRepo := mocks.NewMockAccountRoleRepository(t)
			refreshSessionRepo := mocks.NewMockRefreshSessionRepository(t)
			passwordHasher := mocks.NewMockPasswordHasher(t)
			tokenGenerator := mocks.NewMockTokenGenerator(t)

			tt.mockBehaviour(adapter{
				account:        accountRepo,
				accountRole:    accountRoleRepo,
				refreshSession: refreshSessionRepo,
				passwordHasher: passwordHasher,
				tokenGenerator: tokenGenerator,
			}, acc)

			uc := NewLoginUC(
				accountRepo, accountRoleRepo, refreshSessionRepo, passwordHasher, tokenGenerator, ttl,
			)

			out, err := uc.Execute(context.Background(), tt.input)

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, out.AccessToken)
				assert.NotEmpty(t, out.RefreshToken)
			}
		})
	}
}
