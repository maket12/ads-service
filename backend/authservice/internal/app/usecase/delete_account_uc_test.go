package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/maket12/ads-service/authservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/authservice/internal/app/errs"
	"github.com/maket12/ads-service/authservice/internal/app/usecase"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	"github.com/maket12/ads-service/authservice/internal/domain/port/mocks"
	pkgerrs "github.com/maket12/ads-service/authservice/pkg/errs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeleteAccountUC_Execute(t *testing.T) {
	type adapter struct {
		account          *mocks.MockAccountRepository
		accountRole      *mocks.MockAccountRoleRepository
		accountPublisher *mocks.MockAccountPublisher
	}

	type testCase struct {
		name          string
		input         dto.DeleteAccountInput
		mockBehaviour func(a adapter, acc *model.Account)
		expectErr     error
	}

	email := gofakeit.Email()
	hashedPassword := gofakeit.Password(true, true, true, true, true, 10)

	var tests = []testCase{
		{
			name: "Success",
			mockBehaviour: func(a adapter, acc *model.Account) {
				a.account.EXPECT().
					GetByID(mock.Anything, acc.ID()).
					Return(acc, nil)

				a.accountRole.EXPECT().
					Get(mock.Anything, acc.ID()).
					Return(model.RestoreAccountRole(acc.ID(), model.RoleUser), nil)

				a.account.EXPECT().
					Update(mock.Anything, acc).
					Return(nil)

				a.accountPublisher.EXPECT().
					PublishAccountDelete(mock.Anything, acc.ID()).
					Return(nil)
			},
			expectErr: nil,
		},
		{
			name: "Failure - account not found",
			mockBehaviour: func(a adapter, acc *model.Account) {
				a.account.EXPECT().
					GetByID(mock.Anything, acc.ID()).
					Return(nil, pkgerrs.ErrObjectNotFound)
			},
			expectErr: ucerrs.ErrAccountNotFound,
		},
		{
			name: "Failure - db error on GetByID",
			mockBehaviour: func(a adapter, acc *model.Account) {
				a.account.EXPECT().
					GetByID(mock.Anything, acc.ID()).
					Return(nil, errors.New("db error"))
			},
			expectErr: ucerrs.ErrGetAccountByIDDB,
		},
		{
			name: "Failure - get account role db error",
			mockBehaviour: func(a adapter, acc *model.Account) {
				a.account.EXPECT().
					GetByID(mock.Anything, acc.ID()).
					Return(acc, nil)

				a.accountRole.EXPECT().
					Get(mock.Anything, acc.ID()).
					Return(nil, errors.New("db error"))
			},
			expectErr: ucerrs.ErrGetAccountRoleDB,
		},
		{
			name: "Failure - cannot delete admin",
			mockBehaviour: func(a adapter, acc *model.Account) {
				a.account.EXPECT().
					GetByID(mock.Anything, acc.ID()).
					Return(acc, nil)

				a.accountRole.EXPECT().
					Get(mock.Anything, acc.ID()).
					Return(model.RestoreAccountRole(acc.ID(), model.RoleAdmin), nil)
			},
			expectErr: ucerrs.ErrCannotDelete,
		},
		{
			name: "Failure - account update db error",
			mockBehaviour: func(a adapter, acc *model.Account) {
				a.account.EXPECT().
					GetByID(mock.Anything, acc.ID()).
					Return(acc, nil)

				a.accountRole.EXPECT().
					Get(mock.Anything, acc.ID()).
					Return(model.RestoreAccountRole(acc.ID(), model.RoleUser), nil)

				a.account.EXPECT().
					Update(mock.Anything, acc).
					Return(errors.New("db error"))
			},
			expectErr: ucerrs.ErrUpdateAccountDB,
		},
		{
			name: "Failure - publish event error",
			mockBehaviour: func(a adapter, acc *model.Account) {
				a.account.EXPECT().
					GetByID(mock.Anything, acc.ID()).
					Return(acc, nil)

				a.accountRole.EXPECT().
					Get(mock.Anything, acc.ID()).
					Return(model.RestoreAccountRole(acc.ID(), model.RoleUser), nil)

				a.account.EXPECT().
					Update(mock.Anything, acc).
					Return(nil)

				a.accountPublisher.EXPECT().
					PublishAccountDelete(mock.Anything, acc.ID()).
					Return(errors.New("publish error"))
			},
			expectErr: ucerrs.ErrPublishEvent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc, err := model.NewAccount(email, hashedPassword)
			assert.NoError(t, err)

			tt.input.AccountID = acc.ID()

			accountRepo := mocks.NewMockAccountRepository(t)
			accountRoleRepo := mocks.NewMockAccountRoleRepository(t)
			accountPublisher := mocks.NewMockAccountPublisher(t)

			tt.mockBehaviour(adapter{
				account:          accountRepo,
				accountRole:      accountRoleRepo,
				accountPublisher: accountPublisher,
			}, acc)

			uc := usecase.NewDeleteAccountUC(accountRepo, accountRoleRepo, accountPublisher)

			out, err := uc.Execute(context.Background(), tt.input)

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
