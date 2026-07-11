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

type RefreshSessionUC struct {
	accountRole       port.AccountRoleRepository
	refreshSession    port.RefreshSessionRepository
	tokenGenerator    port.TokenGenerator
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
	// Validate the specified token
	accountID, oldSessionID, err := uc.tokenGenerator.ValidateRefreshToken(
		ctx, in.RefreshToken,
	)
	if err != nil {
		return dto.RefreshSessionOutput{}, ucerrs.ErrInvalidRefreshToken
	}

	// Find the old session
	oldSession, err := uc.refreshSession.GetByID(ctx, oldSessionID)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.RefreshSessionOutput{}, ucerrs.ErrInvalidRefreshToken
		}
		return dto.RefreshSessionOutput{}, ucerrs.Wrap(
			ucerrs.ErrGetRefreshSessionByIDDB, err,
		)
	}

	// =========================================================
	//                     BREACH DETECTION
	// =========================================================
	if err = uc.breachDetection(ctx, oldSession,
		in.RefreshToken,
		in.IP, in.UserAgent,
	); err != nil {
		return dto.RefreshSessionOutput{}, err
	}
	// =========================================================

	// Revoke the old session and save it into the database
	if err = oldSession.RevokeByRotation(); err != nil {
		return dto.RefreshSessionOutput{}, ucerrs.Wrap(
			ucerrs.ErrCannotRevoke, err,
		)
	}

	if err = uc.refreshSession.Update(ctx, oldSession); err != nil {
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
	var sessionID = uuid.New()

	tokensPair, err := uc.tokenGenerator.GeneratePair(
		ctx, accountID, accRole.Role().String(), sessionID,
	)
	if err != nil {
		return dto.RefreshSessionOutput{}, ucerrs.Wrap(
			ucerrs.ErrGenerateTokensPair, err,
		)
	}

	// Create new refresh session
	refreshSession, err := model.NewRefreshSession(
		sessionID, accountID, utils.HashToken(tokensPair.Refresh),
		&oldSessionID, in.IP, in.UserAgent, uc.refreshSessionTTL,
	)
	if err != nil {
		return dto.RefreshSessionOutput{}, ucerrs.Wrap(
			ucerrs.ErrInvalidInput, err,
		)
	}

	if err = uc.refreshSession.Create(ctx, refreshSession); err != nil {
		return dto.RefreshSessionOutput{}, ucerrs.Wrap(
			ucerrs.ErrCreateRefreshSessionDB, err,
		)
	}

	// Output
	return dto.RefreshSessionOutput{
		AccessToken:  tokensPair.Access,
		RefreshToken: tokensPair.Refresh,
	}, nil
}

func (uc *RefreshSessionUC) breachDetection(
	ctx context.Context,
	session *model.RefreshSession,
	token string,
	ip, userAgent *string,
) error {
	// Revoke descendants due to repeated request with the rotated token
	if session.IsRevoked() && session.RevokeReason() != nil &&
		*session.RevokeReason() == model.ReasonTokenRotation {
		_ = uc.refreshSession.RevokeDescendants(
			ctx,
			session.ID(),
			utils.VPtr(model.ReasonCompromisedReuse.String()),
		)
		return ucerrs.ErrInvalidRefreshToken
	}

	// Revoke due to suspicious env
	if !utils.ComparePtr(session.IP(), ip) ||
		!utils.ComparePtr(session.UserAgent(), userAgent) {
		_ = session.RevokeBySuspiciousEnv()
		_ = uc.refreshSession.Update(ctx, session)
		return ucerrs.ErrInvalidRefreshToken
	}

	// Base validation
	if !session.IsActive() {
		return ucerrs.ErrInvalidRefreshToken
	}
	if utils.HashToken(token) != session.RefreshTokenHash() {
		return ucerrs.ErrInvalidRefreshToken
	}

	return nil
}
