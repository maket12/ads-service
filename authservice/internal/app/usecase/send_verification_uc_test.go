package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/maket12/ads-service/authservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/authservice/internal/app/errs"
	"github.com/maket12/ads-service/authservice/internal/app/usecase"
	"github.com/maket12/ads-service/authservice/internal/domain/port/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSendVerificationUC_Execute(t *testing.T) {
	type adapter struct {
		verificationToken *mocks.MockVerificationTokenRepository
		emailSender       *mocks.MockEmailSender
	}

	type testCase struct {
		name          string
		input         dto.SendVerificationInput
		mockBehaviour func(a adapter)
		expectErr     error
	}

	accountID := uuid.New()
	email := "verify@example.com"

	var tests = []testCase{
		{
			name: "Success",
			input: dto.SendVerificationInput{
				AccountID: accountID,
				Email:     email,
			},
			mockBehaviour: func(a adapter) {
				a.verificationToken.EXPECT().Save(mock.Anything, mock.AnythingOfType("*model.VerificationToken")).Return(nil)
				a.emailSender.EXPECT().SendVerificationEmail(mock.Anything, email, mock.AnythingOfType("string")).Return(nil)
			},
			expectErr: nil,
		},
		{
			name: "Failure - db error on token save",
			input: dto.SendVerificationInput{
				AccountID: accountID,
				Email:     email,
			},
			mockBehaviour: func(a adapter) {
				a.verificationToken.EXPECT().Save(mock.Anything, mock.AnythingOfType("*model.VerificationToken")).Return(errors.New("db error"))
			},
			expectErr: ucerrs.ErrSaveVerificationTokenDB,
		},
		{
			name: "Failure - mail delivery service error",
			input: dto.SendVerificationInput{
				AccountID: accountID,
				Email:     email,
			},
			mockBehaviour: func(a adapter) {
				a.verificationToken.EXPECT().Save(mock.Anything, mock.AnythingOfType("*model.VerificationToken")).Return(nil)
				a.emailSender.EXPECT().SendVerificationEmail(mock.Anything, email, mock.AnythingOfType("string")).Return(errors.New("smtp failure"))
			},
			expectErr: ucerrs.ErrSendVerificationEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verificationTokenRepo := mocks.NewMockVerificationTokenRepository(t)
			emailSenderMock := mocks.NewMockEmailSender(t)

			tt.mockBehaviour(adapter{
				verificationToken: verificationTokenRepo,
				emailSender:       emailSenderMock,
			})

			uc := usecase.NewSendVerificationUC(verificationTokenRepo, emailSenderMock, 10*time.Minute)

			out, err := uc.Execute(context.Background(), tt.input)

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
				assert.False(t, out.Sent)
			} else {
				assert.NoError(t, err)
				assert.True(t, out.Sent)
			}
		})
	}
}
