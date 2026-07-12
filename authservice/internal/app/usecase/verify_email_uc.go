package usecase

import (
	"context"
	"errors"

	"github.com/maket12/ads-service/authservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/authservice/internal/app/errs"
	"github.com/maket12/ads-service/authservice/internal/domain/port"
	pkgerrs "github.com/maket12/ads-service/authservice/pkg/errs"
)

type VerifyEmailUC struct {
	account           port.AccountRepository
	verificationToken port.VerificationTokenRepository
	emailSender       port.EmailSender
}

func NewVerifyEmailUC(
	account port.AccountRepository,
	verificationToken port.VerificationTokenRepository,
	emailSender port.EmailSender,
) *VerifyEmailUC {
	return &VerifyEmailUC{
		account:           account,
		verificationToken: verificationToken,
		emailSender:       emailSender,
	}
}

func (uc *VerifyEmailUC) Execute(ctx context.Context, in dto.VerifyEmailInput) (dto.VerifyEmailOutput, error) {
	// Get verification token
	vToken, err := uc.verificationToken.Get(ctx, in.Token)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.VerifyEmailOutput{}, ucerrs.ErrVerificationTokenNotFound
		}
		return dto.VerifyEmailOutput{}, ucerrs.Wrap(
			ucerrs.ErrGetVerificationTokenDB, err,
		)
	}

	// Validation
	if vToken.IsExpired() {
		return dto.VerifyEmailOutput{}, ucerrs.Wrap(
			ucerrs.ErrInvalidInput, errors.New("token is expired"),
		)
	}

	// Find account and update it
	account, err := uc.account.GetByID(ctx, vToken.AccountID())
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.VerifyEmailOutput{}, ucerrs.ErrAccountNotFound
		}
		return dto.VerifyEmailOutput{}, ucerrs.Wrap(
			ucerrs.ErrGetAccountByIDDB, err,
		)
	}

	account.VerifyEmail()

	if err = uc.account.VerifyEmail(ctx, account); err != nil {
	}
}
