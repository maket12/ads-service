package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
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

func TestSendVerificationUC_Execute(t *testing.T) {
	type adapter struct {
		account           *mocks.MockAccountRepository
		verificationToken *mocks.MockVerificationTokenRepository
		emailSender       *mocks.MockEmailSender
	}

	type testCase struct {
		name          string
		input         dto.SendVerificationInput
		mockBehaviour func(a adapter)
		expectErr     error
		expectSent    bool
	}

	accountID := uuid.New()
	email := gofakeit.Email()

	unverifiedAccount, err := model.NewAccount(email, "hashed-pass")
	assert.NoError(t, err)

	verifiedAccount, err := model.NewAccount(email, "hashed-pass")
	assert.NoError(t, err)
	verifiedAccount.VerifyEmail()

	blockedAccount, err := model.NewAccount(email, "hashed-pass")
	assert.NoError(t, err)
	_ = blockedAccount.Block()

	deletedAccount, err := model.NewAccount(email, "hashed-pass")
	assert.NoError(t, err)
	_ = deletedAccount.Delete()

	var tests = []testCase{
		{
			name:  "Success",
			input: dto.SendVerificationInput{AccountID: accountID},
			mockBehaviour: func(a adapter) {
				a.account.EXPECT().GetByID(mock.Anything, accountID).Return(unverifiedAccount, nil)
				a.verificationToken.EXPECT().Save(mock.Anything, mock.AnythingOfType("*model.VerificationToken")).Return(nil)
				a.emailSender.EXPECT().SendVerificationEmail(mock.Anything, email, mock.AnythingOfType("string")).Return(nil)
			},
			expectErr:  nil,
			expectSent: true,
		},
		{
			name:  "Success - already verified, nothing is sent",
			input: dto.SendVerificationInput{AccountID: accountID},
			mockBehaviour: func(a adapter) {
				a.account.EXPECT().GetByID(mock.Anything, accountID).Return(verifiedAccount, nil)
			},
			expectErr:  nil,
			expectSent: false,
		},
		{
			name:  "Failure - account is blocked",
			input: dto.SendVerificationInput{AccountID: accountID},
			mockBehaviour: func(a adapter) {
				a.account.EXPECT().GetByID(mock.Anything, accountID).Return(blockedAccount, nil)
			},
			expectErr:  ucerrs.ErrCannotLogin,
			expectSent: false,
		},
		{
			name:  "Failure - account is deleted",
			input: dto.SendVerificationInput{AccountID: accountID},
			mockBehaviour: func(a adapter) {
				a.account.EXPECT().GetByID(mock.Anything, accountID).Return(deletedAccount, nil)
			},
			expectErr:  ucerrs.ErrCannotLogin,
			expectSent: false,
		},
		{
			name:  "Failure - account not found",
			input: dto.SendVerificationInput{AccountID: accountID},
			mockBehaviour: func(a adapter) {
				a.account.EXPECT().GetByID(mock.Anything, accountID).Return(nil, pkgerrs.ErrObjectNotFound)
			},
			expectErr:  ucerrs.ErrAccountNotFound,
			expectSent: false,
		},
		{
			name:  "Failure - db error on account get",
			input: dto.SendVerificationInput{AccountID: accountID},
			mockBehaviour: func(a adapter) {
				a.account.EXPECT().GetByID(mock.Anything, accountID).Return(nil, errors.New("db error"))
			},
			expectErr:  ucerrs.ErrGetAccountByIDDB,
			expectSent: false,
		},
		{
			name:  "Failure - db error on token save",
			input: dto.SendVerificationInput{AccountID: accountID},
			mockBehaviour: func(a adapter) {
				a.account.EXPECT().GetByID(mock.Anything, accountID).Return(unverifiedAccount, nil)
				a.verificationToken.EXPECT().Save(mock.Anything, mock.AnythingOfType("*model.VerificationToken")).Return(errors.New("db error"))
			},
			expectErr:  ucerrs.ErrSaveVerificationTokenDB,
			expectSent: false,
		},
		{
			name:  "Failure - mail delivery service error",
			input: dto.SendVerificationInput{AccountID: accountID},
			mockBehaviour: func(a adapter) {
				a.account.EXPECT().GetByID(mock.Anything, accountID).Return(unverifiedAccount, nil)
				a.verificationToken.EXPECT().Save(mock.Anything, mock.AnythingOfType("*model.VerificationToken")).Return(nil)
				a.emailSender.EXPECT().SendVerificationEmail(mock.Anything, email, mock.AnythingOfType("string")).Return(errors.New("smtp failure"))
			},
			expectErr:  ucerrs.ErrSendVerificationEmail,
			expectSent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := mocks.NewMockAccountRepository(t)
			verificationTokenRepo := mocks.NewMockVerificationTokenRepository(t)
			emailSenderMock := mocks.NewMockEmailSender(t)

			tt.mockBehaviour(adapter{
				account:           accountRepo,
				verificationToken: verificationTokenRepo,
				emailSender:       emailSenderMock,
			})

			uc := usecase.NewSendVerificationUC(accountRepo, verificationTokenRepo, emailSenderMock, 10*time.Minute)

			out, err := uc.Execute(context.Background(), tt.input)

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectSent, out.Sent)
		})
	}
}
