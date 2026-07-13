package model

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	pkgerrs "github.com/maket12/ads-service/authservice/pkg/errs"
)

const minTTL = time.Minute

// ================ Rich model for Verification Token ================

type VerificationToken struct {
	token     string
	accountID uuid.UUID
	ttl       time.Duration
	expiresAt time.Time
}

func NewVerificationToken(
	accountID uuid.UUID,
	ttl time.Duration,
) (*VerificationToken, error) {
	if accountID == uuid.Nil {
		return nil, pkgerrs.NewValueInvalidError("account_id")
	}
	if ttl < minTTL {
		return nil, pkgerrs.NewValueInvalidError("token_ttl")
	}

	// Generate a random token
	b := make([]byte, 32)
	_, _ = rand.Read(b)

	return &VerificationToken{
		token:     hex.EncodeToString(b),
		accountID: accountID,
		ttl:       ttl,
		expiresAt: time.Now().Add(ttl),
	}, nil
}

func RestoreVerificationToken(
	token string,
	accountID uuid.UUID,
	ttl time.Duration,
	expiresAt time.Time,
) *VerificationToken {
	return &VerificationToken{
		token:     token,
		accountID: accountID,
		ttl:       ttl,
		expiresAt: expiresAt,
	}
}

// ================ Read-Only ================

func (t *VerificationToken) Token() string        { return t.token }
func (t *VerificationToken) AccountID() uuid.UUID { return t.accountID }
func (t *VerificationToken) TTL() time.Duration   { return t.ttl }
func (t *VerificationToken) ExpiresAt() time.Time { return t.expiresAt }

// ================ Business logic ================

func (t *VerificationToken) IsExpired() bool {
	return t.expiresAt.Before(time.Now())
}
