package usecase

import (
	"context"
	"errors"

	"github.com/maket12/ads-service/authservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/authservice/internal/app/errs"
	"github.com/maket12/ads-service/authservice/internal/app/utils"
	"github.com/maket12/ads-service/authservice/internal/domain/port"
	pkgerrs "github.com/maket12/ads-service/pkg/errs"
)

type LogoutUC struct {
	refreshSession port.RefreshSessionRepository
	tokenGenerator port.TokenGenerator
}

func NewLogoutUC(
	refreshSession port.RefreshSessionRepository,
	tokenGenerator port.TokenGenerator,
) *LogoutUC {
	return &LogoutUC{
		refreshSession: refreshSession,
		tokenGenerator: tokenGenerator,
	}
}

func (uc *LogoutUC) Execute(ctx context.Context, in dto.LogoutInput) (dto.LogoutOutput, error) {
	// Find session
	_, oldSessionID, err := uc.tokenGenerator.ValidateRefreshToken(
		ctx, in.RefreshToken,
	)
	if err != nil {
		return dto.LogoutOutput{}, ucerrs.ErrInvalidRefreshToken
	}

	session, err := uc.refreshSession.GetByID(ctx, oldSessionID)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.LogoutOutput{}, ucerrs.ErrInvalidRefreshToken
		}
		return dto.LogoutOutput{}, ucerrs.Wrap(
			ucerrs.ErrGetRefreshSessionByIDDB, err,
		)
	}

	// Validate and revoke
	if !session.IsActive() {
		return dto.LogoutOutput{}, ucerrs.ErrInvalidRefreshToken
	}

	if utils.HashToken(in.RefreshToken) != session.RefreshTokenHash() {
		return dto.LogoutOutput{}, ucerrs.ErrInvalidRefreshToken
	}

	var reason = "logout"
	if err := session.Revoke(&reason); err != nil {
		return dto.LogoutOutput{}, ucerrs.ErrCannotRevoke
	}

	if err := uc.refreshSession.Revoke(ctx, session); err != nil {
		return dto.LogoutOutput{}, ucerrs.Wrap(
			ucerrs.ErrRevokeRefreshSessionDB, err,
		)
	}

	return dto.LogoutOutput{Logout: true}, nil
}
