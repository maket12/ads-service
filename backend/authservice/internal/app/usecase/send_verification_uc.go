package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/maket12/ads-service/authservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/authservice/internal/app/errs"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	"github.com/maket12/ads-service/authservice/internal/domain/port"
	pkgerrs "github.com/maket12/ads-service/authservice/pkg/errs"
)

type SendVerificationUC struct {
	account           port.AccountRepository
	verificationToken port.VerificationTokenRepository
	emailSender       port.EmailSender
	tokenTTL          time.Duration
}

func NewSendVerificationUC(
	account port.AccountRepository,
	verificationToken port.VerificationTokenRepository,
	emailSender port.EmailSender,
	tokenTTL time.Duration,
) *SendVerificationUC {
	return &SendVerificationUC{
		account:           account,
		verificationToken: verificationToken,
		emailSender:       emailSender,
		tokenTTL:          tokenTTL,
	}
}

func (uc *SendVerificationUC) Execute(ctx context.Context, in dto.SendVerificationInput) (dto.SendVerificationOutput, error) {
	// Find account and validate it
	account, err := uc.account.GetByID(ctx, in.AccountID)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.SendVerificationOutput{Sent: false}, ucerrs.ErrAccountNotFound
		}
		return dto.SendVerificationOutput{Sent: false}, ucerrs.Wrap(
			ucerrs.ErrGetAccountByIDDB, err,
		)
	}

	if account.EmailVerified() {
		return dto.SendVerificationOutput{Sent: false}, nil
	}

	if !account.CanLogin() {
		return dto.SendVerificationOutput{}, ucerrs.ErrCannotLogin
	}

	// Create verification token
	vToken, err := model.NewVerificationToken(account.ID(), uc.tokenTTL)
	if err != nil {
		return dto.SendVerificationOutput{Sent: false}, ucerrs.Wrap(
			ucerrs.ErrInvalidInput, err,
		)
	}

	// Save it into database
	if err = uc.verificationToken.Save(ctx, vToken); err != nil {
		return dto.SendVerificationOutput{Sent: false}, ucerrs.Wrap(
			ucerrs.ErrSaveVerificationTokenDB, err,
		)
	}

	// Send it to specified email
	err = uc.emailSender.SendVerificationEmail(ctx, account.Email(), vToken.Token())
	if err != nil {
		return dto.SendVerificationOutput{Sent: false}, ucerrs.Wrap(
			ucerrs.ErrSendVerificationEmail, err,
		)
	}

	return dto.SendVerificationOutput{Sent: true}, nil
}
