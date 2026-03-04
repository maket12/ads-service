package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/maket12/ads-service/authservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/authservice/internal/app/errs"
	"github.com/maket12/ads-service/authservice/internal/app/usecase"
	"github.com/maket12/ads-service/authservice/internal/domain/port/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterUC_Execute(t *testing.T) {
	type adapter struct {
		account          *mocks.AccountRepository
		accountRole      *mocks.AccountRoleRepository
		passwordHasher   *mocks.PasswordHasher
		accountPublisher *mocks.AccountPublisher
	}

	type testCase struct {
		name    string
		input   dto.RegisterInput
		prepare func(a adapter)
		wantErr error
	}

	var tests = []testCase{
		{
			name: "Success",
			input: dto.RegisterInput{
				Email:    "test@example.com",
				Password: "securePassword123",
			},
			prepare: func(a adapter) {
				a.passwordHasher.On("Hash", "securePassword123").
					Return("hashed_password", nil)

				a.account.On("Create", mock.Anything, mock.MatchedBy(func(acc interface{ Email() string }) bool {
					return acc.Email() == "test@example.com"
				})).Return(nil)

				a.accountRole.On("Create", mock.Anything, mock.Anything).
					Return(nil)

				a.accountPublisher.On("PublishAccountCreate", mock.Anything, mock.Anything).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "Error - hashing password",
			input: dto.RegisterInput{
				Email:    "test@example.com",
				Password: "123",
			},
			prepare: func(a adapter) {
				a.passwordHasher.On("Hash", "123").
					Return("", errors.New("salt error"))
			},
			wantErr: ucerrs.ErrHashPassword,
		},
		{
			name: "Error - create account",
			input: dto.RegisterInput{
				Email:    "exists@example.com",
				Password: "password",
			},
			prepare: func(a adapter) {
				a.passwordHasher.On("Hash", "password").
					Return("hashed", nil)
				a.account.On("Create", mock.Anything, mock.Anything).
					Return(errors.New("db error"))
			},
			wantErr: ucerrs.ErrCreateAccountDB,
		},
		{
			name: "Error - create account role",
			input: dto.RegisterInput{
				Email:    "fail@example.com",
				Password: "password",
			},
			prepare: func(a adapter) {
				a.passwordHasher.On("Hash", "password").
					Return("hashed", nil)
				a.account.On("Create", mock.Anything, mock.Anything).
					Return(nil)
				a.accountRole.On("Create", mock.Anything, mock.Anything).
					Return(errors.New("db role error"))
			},
			wantErr: ucerrs.ErrCreateAccountRoleDB,
		},
		{
			name: "Error - create a rabbitmq event",
			input: dto.RegisterInput{
				Email:    "test@example.com",
				Password: "securePassword123",
			},
			prepare: func(a adapter) {
				a.passwordHasher.On("Hash", "securePassword123").
					Return("hashed_password", nil)

				a.account.On("Create", mock.Anything, mock.MatchedBy(func(acc interface{ Email() string }) bool {
					return acc.Email() == "test@example.com"
				})).Return(nil)

				a.accountRole.On("Create", mock.Anything, mock.Anything).
					Return(nil)

				a.accountPublisher.On("PublishAccountCreate", mock.Anything, mock.Anything).
					Return(errors.New("account publisher error"))
			},
			wantErr: ucerrs.ErrPublishEvent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := adapter{
				account:          mocks.NewAccountRepository(t),
				accountRole:      mocks.NewAccountRoleRepository(t),
				passwordHasher:   mocks.NewPasswordHasher(t),
				accountPublisher: mocks.NewAccountPublisher(t),
			}

			if tt.prepare != nil {
				tt.prepare(a)
			}

			uc := usecase.NewRegisterUC(a.account, a.accountRole, a.passwordHasher, a.accountPublisher)

			res, err := uc.Execute(context.Background(), tt.input)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, uuid.Nil, res.AccountID)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, res.AccountID)
			}
		})
	}
}
