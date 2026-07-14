package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"github.com/maket12/ads-service/authservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/authservice/internal/app/errs"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	"github.com/maket12/ads-service/authservice/internal/domain/port"
	pkgerrs "github.com/maket12/ads-service/authservice/pkg/errs"
	"github.com/maket12/ads-service/authservice/pkg/utils"

	"github.com/google/uuid"
)

type LoginUC struct {
	trManager      trm.Manager
	account        port.AccountRepository
	accountRole    port.AccountRoleRepository
	refreshSession port.RefreshSessionRepository
	passwordHasher port.PasswordHasher
	tokenGenerator port.TokenGenerator
	refreshTTL     time.Duration
}

func NewLoginUC(
	trManager trm.Manager,
	account port.AccountRepository,
	accountRole port.AccountRoleRepository,
	refreshSession port.RefreshSessionRepository,
	passwordHasher port.PasswordHasher,
	tokenGenerator port.TokenGenerator,
	refreshTTL time.Duration,
) *LoginUC {
	return &LoginUC{
		trManager:      trManager,
		account:        account,
		accountRole:    accountRole,
		refreshSession: refreshSession,
		passwordHasher: passwordHasher,
		tokenGenerator: tokenGenerator,
		refreshTTL:     refreshTTL,
	}
}

func (uc *LoginUC) Execute(ctx context.Context, in dto.LoginInput) (dto.LoginOutput, error) {
	// Find an account and compare passwords
	account, err := uc.account.GetByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.LoginOutput{}, ucerrs.ErrInvalidCredentials
		}
		return dto.LoginOutput{}, ucerrs.Wrap(
			ucerrs.ErrGetAccountByEmailDB, err,
		)
	}

	if !uc.passwordHasher.Compare(account.PasswordHash(), in.Password) {
		return dto.LoginOutput{}, ucerrs.ErrInvalidCredentials
	}

	// Account validation
	if ok := account.CanLogin(); !ok {
		return dto.LoginOutput{}, ucerrs.ErrCannotLogin
	}

	// Find an account role
	accRole, err := uc.accountRole.Get(ctx, account.ID())
	if err != nil {
		return dto.LoginOutput{}, ucerrs.Wrap(ucerrs.ErrGetAccountRoleDB, err)
	}

	// Generate tokens
	var sessionID = uuid.New()

	tokensPair, err := uc.tokenGenerator.GeneratePair(
		ctx,
		account.ID(),
		accRole.Role().String(),
		sessionID,
	)
	if err != nil {
		return dto.LoginOutput{}, ucerrs.Wrap(
			ucerrs.ErrGenerateTokensPair, err,
		)
	}

	var output dto.LoginOutput

	err = uc.trManager.Do(ctx, func(txCtx context.Context) error {
		// Update account
		updErr := account.MarkLogin()
		if updErr != nil {
			return ucerrs.Wrap(ucerrs.ErrInvalidInput, updErr)
		}

		if updErr = uc.account.Update(ctx, account); updErr != nil {
			return ucerrs.Wrap(ucerrs.ErrUpdateAccountDB, updErr)
		}

		// Revoke all sessions for the same device
		if updErr = uc.refreshSession.RevokeAllForAccountByIPUA(ctx,
			account.ID(), in.IP, in.UserAgent,
			utils.VPtr(model.ReasonReAuth.String()),
		); updErr != nil {
			return ucerrs.Wrap(ucerrs.ErrRevokeAllForAccountByIPUADB, updErr)
		}

		// Create a refresh session and save it into database
		refreshSession, createErr := model.NewRefreshSession(
			sessionID, account.ID(), utils.HashToken(tokensPair.Refresh),
			nil, in.IP, in.UserAgent, uc.refreshTTL,
		)
		if createErr != nil {
			return ucerrs.Wrap(ucerrs.ErrInvalidInput, createErr)
		}

		if createErr = uc.refreshSession.Create(ctx, refreshSession); createErr != nil {
			return ucerrs.Wrap(ucerrs.ErrCreateRefreshSessionDB, createErr)
		}

		output = dto.LoginOutput{
			AccessToken:  tokensPair.Access,
			RefreshToken: tokensPair.Refresh,
		}

		return nil
	})
	if err != nil {
		return dto.LoginOutput{}, err
	}

	return output, nil
}
