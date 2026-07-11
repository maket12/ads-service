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
	"github.com/maket12/ads-service/authservice/pkg/utils"

	"github.com/google/uuid"
)

type LoginUC struct {
	account        port.AccountRepository
	accountRole    port.AccountRoleRepository
	refreshSession port.RefreshSessionRepository
	passwordHasher port.PasswordHasher
	tokenGenerator port.TokenGenerator
	refreshTTL     time.Duration
}

func NewLoginUC(
	account port.AccountRepository,
	accountRole port.AccountRoleRepository,
	refreshSession port.RefreshSessionRepository,
	passwordHasher port.PasswordHasher,
	tokenGenerator port.TokenGenerator,
	refreshTTL time.Duration,
) *LoginUC {
	return &LoginUC{
		account:        account,
		accountRole:    accountRole,
		refreshSession: refreshSession,
		passwordHasher: passwordHasher,
		tokenGenerator: tokenGenerator,
		refreshTTL:     refreshTTL,
	}
}

func (uc *LoginUC) Execute(ctx context.Context, in dto.LoginInput) (dto.LoginOutput, error) {
	// Find an account
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

	// Update account
	if err = account.MarkLogin(); err != nil {
		return dto.LoginOutput{}, ucerrs.Wrap(
			ucerrs.ErrInvalidInput, err,
		)
	}

	if err = uc.account.MarkLogin(ctx, account); err != nil {
		return dto.LoginOutput{}, ucerrs.Wrap(
			ucerrs.ErrUpdateAccountDB, err,
		)
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

	// Create a refresh session
	refreshSession, err := model.NewRefreshSession(
		sessionID, account.ID(), utils.HashToken(tokensPair.Refresh),
		nil, in.IP, in.UserAgent, uc.refreshTTL,
	)
	if err != nil {
		return dto.LoginOutput{}, ucerrs.Wrap(
			ucerrs.ErrInvalidInput, err,
		)
	}

	if err = uc.refreshSession.Create(ctx, refreshSession); err != nil {
		return dto.LoginOutput{}, ucerrs.Wrap(
			ucerrs.ErrCreateRefreshSessionDB, err,
		)
	}

	// Output
	return dto.LoginOutput{
		AccessToken:  tokensPair.Access,
		RefreshToken: tokensPair.Refresh,
	}, nil
}
