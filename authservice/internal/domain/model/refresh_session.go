package model

import (
	"errors"
	pkgerrs "github.com/maket12/ads-service/pkg/errs"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTokenAlreadyRevoked = errors.New("token has been revoked earlier")
)

// ================ Rich model for Refresh Session ================

type RefreshSession struct {
	id               uuid.UUID
	accountID        uuid.UUID
	refreshTokenHash string
	createdAt        time.Time
	expiresAt        time.Time
	revokedAt        *time.Time
	revokeReason     *string
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
		rotatedFrom:      rotatedFrom,
		ip:               ip,
		userAgent:        userAgent,
	}, nil
}

func RestoreRefreshSession(
	id, accountID uuid.UUID, refreshTokenHash string,
	createdAt, expiresAt time.Time, revokedAt *time.Time,
	revokeReason *string, rotatedFrom *uuid.UUID,
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

func (r *RefreshSession) ID() uuid.UUID            { return r.id }
func (r *RefreshSession) AccountID() uuid.UUID     { return r.accountID }
func (r *RefreshSession) RefreshTokenHash() string { return r.refreshTokenHash }
func (r *RefreshSession) CreatedAt() time.Time     { return r.createdAt }
func (r *RefreshSession) ExpiresAt() time.Time     { return r.expiresAt }
func (r *RefreshSession) RevokedAt() *time.Time    { return r.revokedAt }
func (r *RefreshSession) RevokeReason() *string    { return r.revokeReason }
func (r *RefreshSession) RotatedFrom() *uuid.UUID  { return r.rotatedFrom }
func (r *RefreshSession) IP() *string              { return r.ip }
func (r *RefreshSession) UserAgent() *string       { return r.userAgent }

func (r *RefreshSession) IsActive() bool  { return !r.IsExpired() && !r.IsRevoked() }
func (r *RefreshSession) IsExpired() bool { return time.Now().After(r.ExpiresAt()) }
func (r *RefreshSession) IsRevoked() bool { return r.RevokedAt() != nil }

// ================ Mutation ================

func (r *RefreshSession) Revoke(reason *string) error {
	if r.IsRevoked() {
		return ErrTokenAlreadyRevoked
	}

	if reason != nil && *reason == "" {
		return pkgerrs.NewValueInvalidError("revoke_reason")
	}

	now := time.Now()
	r.revokedAt = &now
	r.revokeReason = reason
	return nil
}
