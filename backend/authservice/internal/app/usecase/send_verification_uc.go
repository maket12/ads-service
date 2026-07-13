package usecase

import (
	"context"
	"time"

	"github.com/maket12/ads-service/authservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/authservice/internal/app/errs"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	"github.com/maket12/ads-service/authservice/internal/domain/port"
)

type SendVerificationUC struct {
	verificationToken port.VerificationTokenRepository
	emailSender       port.EmailSender
	tokenTTL          time.Duration
}

func NewSendVerificationUC(
	verificationToken port.VerificationTokenRepository,
	emailSender port.EmailSender,
	tokenTTL time.Duration,
) *SendVerificationUC {
	return &SendVerificationUC{
		verificationToken: verificationToken,
		emailSender:       emailSender,
		tokenTTL:          tokenTTL,
	}
}

func (uc *SendVerificationUC) Execute(ctx context.Context, in dto.SendVerificationInput) (dto.SendVerificationOutput, error) {
	// Create verification token
	vToken, err := model.NewVerificationToken(in.AccountID, uc.tokenTTL)
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
	err = uc.emailSender.SendVerificationEmail(ctx, in.Email, vToken.Token())
	if err != nil {
		return dto.SendVerificationOutput{Sent: false}, ucerrs.Wrap(
			ucerrs.ErrSendVerificationEmail, err,
		)
	}

	return dto.SendVerificationOutput{Sent: true}, nil
}
