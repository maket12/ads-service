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
			return dto.VerifyEmailOutput{Verified: false}, ucerrs.ErrVerificationTokenNotFound
		}
		return dto.VerifyEmailOutput{Verified: false}, ucerrs.Wrap(
			ucerrs.ErrGetVerificationTokenDB, err,
		)
	}

	// Validation
	if vToken.IsExpired() {
		return dto.VerifyEmailOutput{Verified: false}, ucerrs.ErrCannotVerify
	}

	// Find account and update it
	account, err := uc.account.GetByID(ctx, vToken.AccountID())
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.VerifyEmailOutput{Verified: false}, ucerrs.ErrAccountNotFound
		}
		return dto.VerifyEmailOutput{Verified: false}, ucerrs.Wrap(
			ucerrs.ErrGetAccountByIDDB, err,
		)
	}

	account.VerifyEmail()

	if err = uc.account.Update(ctx, account); err != nil {
		return dto.VerifyEmailOutput{Verified: false}, ucerrs.Wrap(
			ucerrs.ErrUpdateAccountDB, err,
		)
	}

	// Delete token (if it still exists)
	_ = uc.verificationToken.Delete(ctx, vToken.Token())

	// Output
	return dto.VerifyEmailOutput{Verified: true}, nil
}
