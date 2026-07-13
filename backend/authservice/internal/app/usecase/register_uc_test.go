package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/maket12/ads-service/authservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/authservice/internal/app/errs"
	"github.com/maket12/ads-service/authservice/internal/app/usecase"
	"github.com/maket12/ads-service/authservice/internal/domain/port/mocks"
	pkgerrs "github.com/maket12/ads-service/authservice/pkg/errs"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterUC_Execute(t *testing.T) {
	type adapter struct {
		account          *mocks.MockAccountRepository
		accountRole      *mocks.MockAccountRoleRepository
		passwordHasher   *mocks.MockPasswordHasher
		accountPublisher *mocks.MockAccountPublisher
	}

	type testCase struct {
		name          string
		input         dto.RegisterInput
		mockBehaviour func(a adapter)
		expectErr     error
	}

	email := "newuser@example.com"
	password := "plain-password"
	hashedPassword := "hashed-password"

	var tests = []testCase{
		{
			name: "Success",
			input: dto.RegisterInput{
				Email:    email,
				Password: password,
			},
			mockBehaviour: func(a adapter) {
				a.passwordHasher.EXPECT().
					Hash(password).
					Return(hashedPassword, nil)

				a.account.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*model.Account")).
					Return(nil)

				a.accountRole.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*model.AccountRole")).
					Return(nil)

				a.accountPublisher.EXPECT().
					PublishAccountCreate(mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(nil)
			},
			expectErr: nil,
		},
		{
			name: "Failure - password hashing error",
			input: dto.RegisterInput{
				Email:    email,
				Password: password,
			},
			mockBehaviour: func(a adapter) {
				a.passwordHasher.EXPECT().
					Hash(password).
					Return("", errors.New("hashing failed"))
			},
			expectErr: ucerrs.ErrHashPassword,
		},
		{
			name: "Failure - account already exists",
			input: dto.RegisterInput{
				Email:    email,
				Password: password,
			},
			mockBehaviour: func(a adapter) {
				a.passwordHasher.EXPECT().
					Hash(password).
					Return(hashedPassword, nil)

				// Благодаря вашему фиксу, теперь эта ошибка корректно обрабатывается
				a.account.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*model.Account")).
					Return(pkgerrs.ErrObjectAlreadyExists)
			},
			expectErr: ucerrs.ErrAccountAlreadyExists,
		},
		{
			name: "Failure - db error on account creation",
			input: dto.RegisterInput{
				Email:    email,
				Password: password,
			},
			mockBehaviour: func(a adapter) {
				a.passwordHasher.EXPECT().
					Hash(password).
					Return(hashedPassword, nil)

				a.account.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*model.Account")).
					Return(errors.New("generic db error"))
			},
			expectErr: ucerrs.ErrCreateAccountDB,
		},
		{
			name: "Failure - db error on account role creation",
			input: dto.RegisterInput{
				Email:    email,
				Password: password,
			},
			mockBehaviour: func(a adapter) {
				a.passwordHasher.EXPECT().
					Hash(password).
					Return(hashedPassword, nil)

				a.account.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*model.Account")).
					Return(nil)

				a.accountRole.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*model.AccountRole")).
					Return(errors.New("role db error"))
			},
			expectErr: ucerrs.ErrCreateAccountRoleDB,
		},
		{
			name: "Failure - event publishing error",
			input: dto.RegisterInput{
				Email:    email,
				Password: password,
			},
			mockBehaviour: func(a adapter) {
				a.passwordHasher.EXPECT().
					Hash(password).
					Return(hashedPassword, nil)

				a.account.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*model.Account")).
					Return(nil)

				a.accountRole.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*model.AccountRole")).
					Return(nil)

				a.accountPublisher.EXPECT().
					PublishAccountCreate(mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(errors.New("rabbitmq unavailable"))
			},
			expectErr: ucerrs.ErrPublishEvent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := mocks.NewMockAccountRepository(t)
			accountRoleRepo := mocks.NewMockAccountRoleRepository(t)
			passwordHasher := mocks.NewMockPasswordHasher(t)
			accountPublisher := mocks.NewMockAccountPublisher(t)

			tt.mockBehaviour(adapter{
				account:          accountRepo,
				accountRole:      accountRoleRepo,
				passwordHasher:   passwordHasher,
				accountPublisher: accountPublisher,
			})

			txManager := mocks.FakeTxManager{}

			uc := usecase.NewRegisterUC(
				txManager, accountRepo, accountRoleRepo, passwordHasher, accountPublisher,
			)

			out, err := uc.Execute(context.Background(), tt.input)

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
				assert.Equal(t, uuid.Nil, out.AccountID)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, out.AccountID)
			}
		})
	}
}
