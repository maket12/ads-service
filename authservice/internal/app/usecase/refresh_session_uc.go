package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/maket12/ads-service/authservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/authservice/internal/app/errs"
	"github.com/maket12/ads-service/authservice/internal/app/utils"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	"github.com/maket12/ads-service/authservice/internal/domain/port"
	pkgerrs "github.com/maket12/ads-service/pkg/errs"

	"github.com/google/uuid"
)

type RefreshSessionUC struct {
	accountRole    port.AccountRoleRepository
	refreshSession port.RefreshSessionRepository
	tokenGenerator port.TokenGenerator

	refreshSessionTTL time.Duration
}

func NewRefreshSessionUC(
	accountRole port.AccountRoleRepository,
	refreshSession port.RefreshSessionRepository,
	tokenGenerator port.TokenGenerator,
	refreshSessionTTL time.Duration,
) *RefreshSessionUC {
	return &RefreshSessionUC{
		accountRole:       accountRole,
		refreshSession:    refreshSession,
		tokenGenerator:    tokenGenerator,
		refreshSessionTTL: refreshSessionTTL,
	}
}

func (uc *RefreshSessionUC) Execute(ctx context.Context, in dto.RefreshSessionInput) (dto.RefreshSessionOutput, error) {
	// Find old session
	accountID, oldSessionID, err := uc.tokenGenerator.ValidateRefreshToken(
		ctx, in.OldRefreshToken,
	)
	if err != nil {
		return dto.RefreshSessionOutput{}, ucerrs.ErrInvalidRefreshToken
	}

	oldSession, err := uc.refreshSession.GetByID(ctx, oldSessionID)

	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.RefreshSessionOutput{}, ucerrs.ErrInvalidRefreshToken
		}
		return dto.RefreshSessionOutput{}, ucerrs.Wrap(
			ucerrs.ErrGetRefreshSessionByIDDB, err,
		)
	}

	// Validate and revoke
	if !oldSession.IsActive() ||
		!utils.ComparePtr(oldSession.IP(), in.IP) ||
		!utils.ComparePtr(oldSession.UserAgent(), in.UserAgent) {
		return dto.RefreshSessionOutput{}, ucerrs.ErrInvalidRefreshToken
	}

	if utils.HashToken(in.OldRefreshToken) != oldSession.RefreshTokenHash() {
		return dto.RefreshSessionOutput{}, ucerrs.ErrInvalidRefreshToken
	}

	var reason = "token rotation"
	if err := oldSession.Revoke(&reason); err != nil {
		return dto.RefreshSessionOutput{}, ucerrs.Wrap(
			ucerrs.ErrCannotRevoke, err,
		)
	}

	if err := uc.refreshSession.Revoke(ctx, oldSession); err != nil {
		return dto.RefreshSessionOutput{}, ucerrs.Wrap(
			ucerrs.ErrRevokeRefreshSessionDB, err,
		)
	}

	// Get account role
	accRole, err := uc.accountRole.Get(ctx, accountID)
	if err != nil {
		return dto.RefreshSessionOutput{}, ucerrs.Wrap(
			ucerrs.ErrGetAccountRoleDB, err,
		)
	}

	// Generate new tokens
	accessToken, err := uc.tokenGenerator.GenerateAccessToken(
		ctx, accountID, accRole.Role().String(),
	)
	if err != nil {
		return dto.RefreshSessionOutput{}, ucerrs.Wrap(
			ucerrs.ErrGenerateAccessToken, err,
		)
	}

	var sessionID = uuid.New()
	refreshToken, err := uc.tokenGenerator.GenerateRefreshToken(
		ctx, accountID, sessionID,
	)
	if err != nil {
		return dto.RefreshSessionOutput{}, ucerrs.Wrap(
			ucerrs.ErrGenerateRefreshToken, err,
		)
	}

	hashedRefreshToken := utils.HashToken(refreshToken)

	// Create new refresh session with rotation
	refreshSession, err := model.NewRefreshSession(
		sessionID, accountID, hashedRefreshToken, &oldSessionID,
		in.IP, in.UserAgent, uc.refreshSessionTTL,
	)
	if err != nil {
		return dto.RefreshSessionOutput{}, ucerrs.Wrap(
			ucerrs.ErrInvalidInput, err,
		)
	}

	if err := uc.refreshSession.Create(ctx, refreshSession); err != nil {
		return dto.RefreshSessionOutput{}, ucerrs.Wrap(
			ucerrs.ErrCreateRefreshSessionDB, err,
		)
	}

	// Output
	return dto.RefreshSessionOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
