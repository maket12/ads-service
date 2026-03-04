package usecase

import (
	"ads/authservice/internal/app/dto"
	"ads/authservice/internal/app/uc_errors"
	"ads/authservice/internal/app/utils"
	"ads/authservice/internal/domain/model"
	"ads/authservice/internal/domain/port"
	"ads/pkg/errs"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type LoginUC struct {
	account        port.AccountRepository
	accountRole    port.AccountRoleRepository
	refreshSession port.RefreshSessionRepository
	passwordHasher port.PasswordHasher
	tokenGenerator port.TokenGenerator

	refreshSessionTTL time.Duration
}

func NewLoginUC(
	account port.AccountRepository,
	accountRole port.AccountRoleRepository,
	refreshSession port.RefreshSessionRepository,
	passwordHasher port.PasswordHasher,
	tokenGenerator port.TokenGenerator,
	refreshSessionTTL time.Duration,
) *LoginUC {
	return &LoginUC{
		account:           account,
		accountRole:       accountRole,
		refreshSession:    refreshSession,
		passwordHasher:    passwordHasher,
		tokenGenerator:    tokenGenerator,
		refreshSessionTTL: refreshSessionTTL,
	}
}

func (uc *LoginUC) Execute(ctx context.Context, in dto.LoginInput) (dto.LoginOutput, error) {
	// Find account
	account, err := uc.account.GetByEmail(ctx, in.Email)

	if err != nil {
		if errors.Is(err, errs.ErrObjectNotFound) {
			return dto.LoginOutput{}, uc_errors.ErrInvalidCredentials
		}
		return dto.LoginOutput{}, uc_errors.Wrap(
			uc_errors.ErrGetAccountByEmailDB, err,
		)
	}

	if !uc.passwordHasher.Compare(account.PasswordHash(), in.Password) {
		return dto.LoginOutput{}, uc_errors.ErrInvalidCredentials
	}

	// Account validation
	if ok := account.CanLogin(); !ok {
		return dto.LoginOutput{}, uc_errors.ErrCannotLogin
	}

	// Update account
	account.MarkLogin()
	if err := uc.account.MarkLogin(ctx, account); err != nil {
		return dto.LoginOutput{}, uc_errors.Wrap(
			uc_errors.ErrUpdateAccountDB, err,
		)
	}

	// Find an account role
	accRole, err := uc.accountRole.Get(ctx, account.ID())
	if err != nil {
		return dto.LoginOutput{}, uc_errors.Wrap(uc_errors.ErrGetAccountRoleDB, err)
	}

	// Generate tokens
	accessToken, err := uc.tokenGenerator.GenerateAccessToken(
		ctx, account.ID(), accRole.Role().String(),
	)
	if err != nil {
		return dto.LoginOutput{}, uc_errors.Wrap(
			uc_errors.ErrGenerateAccessToken, err,
		)
	}

	var sessionID = uuid.New()
	refreshToken, err := uc.tokenGenerator.GenerateRefreshToken(
		ctx, account.ID(), sessionID,
	)
	if err != nil {
		return dto.LoginOutput{}, uc_errors.Wrap(
			uc_errors.ErrGenerateRefreshToken, err,
		)
	}

	hashedRefreshToken := utils.HashToken(refreshToken)

	// Create refresh session
	refreshSession, err := model.NewRefreshSession(
		sessionID, account.ID(), hashedRefreshToken, nil,
		in.IP, in.UserAgent, uc.refreshSessionTTL,
	)
	if err != nil {
		return dto.LoginOutput{}, uc_errors.Wrap(
			uc_errors.ErrInvalidInput, err,
		)
	}

	if err := uc.refreshSession.Create(ctx, refreshSession); err != nil {
		return dto.LoginOutput{}, uc_errors.Wrap(
			uc_errors.ErrCreateRefreshSessionDB, err,
		)
	}

	// Output
	return dto.LoginOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
