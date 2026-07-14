package model

import (
	"errors"
	"time"

	pkgerrs "github.com/maket12/ads-service/authservice/pkg/errs"

	"github.com/google/uuid"
)

var (
	ErrCannotRevokeToken = errors.New("token has been revoked earlier")
)

type RevokeReason string

const (
	ReasonLogout           RevokeReason = "logout"
	ReasonTokenRotation    RevokeReason = "token rotation"
	ReasonSuspiciousEnv    RevokeReason = "suspicious environment change"
	ReasonCompromisedReuse RevokeReason = "compromised: reuse of rotated token"
	ReasonRoleChanged      RevokeReason = "role changed"
	ReasonReAuth           RevokeReason = "reauthenticated"
)

func (r RevokeReason) String() string { return string(r) }

// ================ Rich model for Refresh Session ================

type RefreshSession struct {
	id               uuid.UUID
	accountID        uuid.UUID
	refreshTokenHash string
	createdAt        time.Time
	expiresAt        time.Time
	revokedAt        *time.Time
	revokeReason     *RevokeReason
	rotatedFrom      *uuid.UUID
	ip               *string
	userAgent        *string
}

func NewRefreshSession(
	id, accountID uuid.UUID, refreshTokenHash string,
	rotatedFrom *uuid.UUID, ip *string, userAgent *string,
	ttl time.Duration,
) (*RefreshSession, error) {
	if id == uuid.Nil {
		return nil, pkgerrs.NewValueInvalidError("session_id")
	}
	if accountID == uuid.Nil {
		return nil, pkgerrs.NewValueInvalidError("account_id")
	}
	if refreshTokenHash == "" {
		return nil, pkgerrs.NewValueRequiredError("refresh_token_hash")
	}
	if rotatedFrom != nil && *rotatedFrom == uuid.Nil {
		return nil, pkgerrs.NewValueInvalidError("rotated_from")
	}
	if ttl <= 0 {
		return nil, pkgerrs.NewValueInvalidError("ttl")
	}

	now := time.Now()
	expiresAt := now.Add(ttl)

	return &RefreshSession{
		id:               id,
		accountID:        accountID,
		refreshTokenHash: refreshTokenHash,
		createdAt:        now,
		expiresAt:        expiresAt,
		revokedAt:        nil,
		revokeReason:     nil,
		rotatedFrom:      rotatedFrom,
		ip:               ip,
		userAgent:        userAgent,
	}, nil
}

func RestoreRefreshSession(
	id, accountID uuid.UUID, refreshTokenHash string,
	createdAt, expiresAt time.Time, revokedAt *time.Time,
	revokeReason *RevokeReason, rotatedFrom *uuid.UUID,
	ip *string, userAgent *string,
) *RefreshSession {
	return &RefreshSession{
		id:               id,
		accountID:        accountID,
		refreshTokenHash: refreshTokenHash,
		createdAt:        createdAt,
		expiresAt:        expiresAt,
		revokedAt:        revokedAt,
		revokeReason:     revokeReason,
		rotatedFrom:      rotatedFrom,
		ip:               ip,
		userAgent:        userAgent,
	}
}

// ================ Read-Only ================

func (r *RefreshSession) ID() uuid.UUID               { return r.id }
func (r *RefreshSession) AccountID() uuid.UUID        { return r.accountID }
func (r *RefreshSession) RefreshTokenHash() string    { return r.refreshTokenHash }
func (r *RefreshSession) CreatedAt() time.Time        { return r.createdAt }
func (r *RefreshSession) ExpiresAt() time.Time        { return r.expiresAt }
func (r *RefreshSession) RevokedAt() *time.Time       { return r.revokedAt }
func (r *RefreshSession) RevokeReason() *RevokeReason { return r.revokeReason }
func (r *RefreshSession) RotatedFrom() *uuid.UUID     { return r.rotatedFrom }
func (r *RefreshSession) IP() *string                 { return r.ip }
func (r *RefreshSession) UserAgent() *string          { return r.userAgent }

// ================ Business logic ================

func (r *RefreshSession) IsActive() bool  { return !r.IsExpired() && !r.IsRevoked() }
func (r *RefreshSession) IsExpired() bool { return time.Now().After(r.ExpiresAt()) }
func (r *RefreshSession) IsRevoked() bool { return r.RevokedAt() != nil }

// ================ Mutation ================

func (r *RefreshSession) revoke(reason RevokeReason) error {
	if r.IsRevoked() {
		return ErrCannotRevokeToken
	}

	now := time.Now()
	r.revokedAt = &now
	r.revokeReason = &reason

	return nil
}

func (r *RefreshSession) RevokeByLogout() error {
	return r.revoke(ReasonLogout)
}

func (r *RefreshSession) RevokeByRotation() error {
	return r.revoke(ReasonTokenRotation)
}

func (r *RefreshSession) RevokeBySuspiciousEnv() error {
	return r.revoke(ReasonSuspiciousEnv)
}

func (r *RefreshSession) RevokeByReAuth() error {
	return r.revoke(ReasonReAuth)
}
