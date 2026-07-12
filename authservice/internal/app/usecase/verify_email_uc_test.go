package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

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

func TestVerifyEmailUC_Execute(t *testing.T) {
	type adapter struct {
		account           *mocks.MockAccountRepository
		verificationToken *mocks.MockVerificationTokenRepository
		emailSender       *mocks.MockEmailSender
	}

	type testCase struct {
		name          string
		input         dto.VerifyEmailInput
		mockBehaviour func(a adapter, account *model.Account, vToken *model.VerificationToken)
		expectErr     error
	}

	rawTokenStr := "token-uuid-string"

	var tests = []testCase{
		{
			name: "Success",
			input: dto.VerifyEmailInput{
				Token: rawTokenStr,
			},
			mockBehaviour: func(a adapter, account *model.Account, vToken *model.VerificationToken) {
				a.verificationToken.EXPECT().Get(mock.Anything, rawTokenStr).Return(vToken, nil)
				a.account.EXPECT().GetByID(mock.Anything, vToken.AccountID()).Return(account, nil)
				a.account.EXPECT().Update(mock.Anything, account).Return(nil)
				a.verificationToken.EXPECT().Delete(mock.Anything, vToken.Token()).Return(nil)
			},
			expectErr: nil,
		},
		{
			name: "Failure - lookup returns token not found",
			input: dto.VerifyEmailInput{
				Token: rawTokenStr,
			},
			mockBehaviour: func(a adapter, account *model.Account, vToken *model.VerificationToken) {
				a.verificationToken.EXPECT().Get(mock.Anything, rawTokenStr).Return(nil, pkgerrs.ErrObjectNotFound)
			},
			expectErr: ucerrs.ErrVerificationTokenNotFound,
		},
		{
			name: "Failure - targeted account records missing",
			input: dto.VerifyEmailInput{
				Token: rawTokenStr,
			},
			mockBehaviour: func(a adapter, account *model.Account, vToken *model.VerificationToken) {
				a.verificationToken.EXPECT().Get(mock.Anything, rawTokenStr).Return(vToken, nil)
				a.account.EXPECT().GetByID(mock.Anything, vToken.AccountID()).Return(nil, pkgerrs.ErrObjectNotFound)
			},
			expectErr: ucerrs.ErrAccountNotFound,
		},
		{
			name: "Failure - persist update database error",
			input: dto.VerifyEmailInput{
				Token: rawTokenStr,
			},
			mockBehaviour: func(a adapter, account *model.Account, vToken *model.VerificationToken) {
				a.verificationToken.EXPECT().Get(mock.Anything, rawTokenStr).Return(vToken, nil)
				a.account.EXPECT().GetByID(mock.Anything, vToken.AccountID()).Return(account, nil)
				a.account.EXPECT().Update(mock.Anything, account).Return(errors.New("db error"))
			},
			expectErr: ucerrs.ErrUpdateAccountDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountID := uuid.New()
			account, err := model.NewAccount("confirm@example.com", "secret")
			assert.NoError(t, err)

			vToken, err := model.NewVerificationToken(accountID, 30*time.Minute)
			assert.NoError(t, err)

			accountRepo := mocks.NewMockAccountRepository(t)
			verificationTokenRepo := mocks.NewMockVerificationTokenRepository(t)
			emailSenderMock := mocks.NewMockEmailSender(t)

			tt.mockBehaviour(adapter{
				account:           accountRepo,
				verificationToken: verificationTokenRepo,
				emailSender:       emailSenderMock,
			}, account, vToken)

			uc := usecase.NewVerifyEmailUC(accountRepo, verificationTokenRepo, emailSenderMock)

			out, err := uc.Execute(context.Background(), tt.input)

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
				assert.False(t, out.Verified)
			} else {
				assert.NoError(t, err)
				assert.True(t, out.Verified)
			}
		})
	}
}
