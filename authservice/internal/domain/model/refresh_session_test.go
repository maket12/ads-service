package model_test

import (
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	pkgerrs "github.com/maket12/ads-service/pkg/errs"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper for taking the pointer
func vPtr[T any](v T) *T {
	return &v
}

func TestNewRefreshSession(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		id          uuid.UUID
		accountID   uuid.UUID
		tokenHash   string
		rotatedFrom *uuid.UUID
		ip          *string
		userAgent   *string
		ttl         time.Duration
		expect      error
	}

	var tests = []testCase{
		{
			name:      "success",
			id:        uuid.New(),
			accountID: uuid.New(),
			tokenHash: "hashed",
			ttl:       time.Minute,
			expect:    nil,
		},
		{
			name:      "nullable session id",
			id:        uuid.Nil,
			accountID: uuid.New(),
			tokenHash: "hashed",
			ttl:       time.Minute,
			expect:    pkgerrs.ErrValueIsInvalid,
		},
		{
			name:      "nullable account id",
			id:        uuid.New(),
			accountID: uuid.Nil,
			tokenHash: "hashed",
			ttl:       time.Minute,
			expect:    pkgerrs.ErrValueIsInvalid,
		},
		{
			name:      "empty token hash",
			id:        uuid.New(),
			accountID: uuid.New(),
			tokenHash: "",
			ttl:       time.Minute,
			expect:    pkgerrs.ErrValueIsRequired,
		},
		{
			name:        "nullable rotated from",
			id:          uuid.New(),
			accountID:   uuid.New(),
			tokenHash:   "hashed",
			rotatedFrom: &uuid.Nil,
			ttl:         time.Minute,
			expect:      pkgerrs.ErrValueIsInvalid,
		},
		{
			name:      "invalid ttl (not positive)",
			id:        uuid.New(),
			accountID: uuid.New(),
			tokenHash: "hashed",
			ip:        vPtr("123.021.234.0"),
			userAgent: vPtr("Mozilla/5.0"),
			ttl:       time.Minute * -1,
			expect:    pkgerrs.ErrValueIsInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := model.NewRefreshSession(
				tt.id, tt.accountID, tt.tokenHash,
				tt.rotatedFrom, tt.ip, tt.userAgent, tt.ttl)
			if tt.expect == nil {
				require.NoError(t, err)
				require.NotNil(t, session)
				assert.NotNil(t, session.ID())
				assert.Equal(t, tt.accountID, session.AccountID())
				assert.Equal(t, tt.tokenHash, session.RefreshTokenHash())
				assert.NotEqual(t, session.CreatedAt(), session.ExpiresAt())
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expect)
				assert.Nil(t, session)
			}
		})
	}
}

func TestRefreshSession_IsExpired(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name      string
		expiresAt time.Time
		expect    bool
	}

	var tests = []testCase{
		{
			name:      "session is expired - true",
			expiresAt: time.Now().Add(time.Hour * -1),
			expect:    true,
		},
		{
			name:      "session is not expired yet - false",
			expiresAt: time.Now().Add(time.Hour),
			expect:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := model.RestoreRefreshSession(
				uuid.New(), uuid.New(), "hashed",
				time.Now(), tt.expiresAt, nil, nil,
				nil, nil, nil)
			assert.Equal(t, tt.expect, session.IsExpired())
		})
	}
}

func TestRefreshSession_IsRevoked(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name      string
		revokedAt *time.Time
		expect    bool
	}

	var tests = []testCase{
		{
			name:      "session is revoked - true",
			revokedAt: vPtr(time.Now()),
			expect:    true,
		},
		{
			name:      "session is not revoked yet - false",
			revokedAt: nil,
			expect:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := model.RestoreRefreshSession(
				uuid.New(), uuid.New(), "hashed",
				time.Now(), time.Now().Add(time.Hour),
				tt.revokedAt, nil,
				nil, nil, nil)
			assert.Equal(t, tt.expect, session.IsRevoked())
		})
	}
}

func TestRefreshSession_IsActive(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name      string
		expiresAt time.Time
		revokedAt *time.Time
		expect    bool
	}

	var tests = []testCase{
		{
			name:      "session is active - true",
			expiresAt: time.Now().Add(time.Hour),
			revokedAt: nil,
			expect:    true,
		},
		{
			name:      "session is expired - false",
			expiresAt: time.Now().Add(time.Hour * -1),
			revokedAt: nil,
			expect:    false,
		},
		{
			name:      "session is not expired, but revoked - false",
			expiresAt: time.Now().Add(time.Hour * 1),
			revokedAt: vPtr(time.Now().Add(time.Minute * -1)),
			expect:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := model.RestoreRefreshSession(
				uuid.New(), uuid.New(), "hashed",
				time.Now(), tt.expiresAt, tt.revokedAt, nil,
				nil, nil, nil)
			assert.Equal(t, tt.expect, session.IsActive())
		})
	}
}

func TestRefreshSession_Revoke(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name      string
		revokedAt *time.Time
		reason    *string
		expect    error
	}

	var tests = []testCase{
		{
			name:      "success",
			revokedAt: nil,
			reason:    vPtr("tests"),
			expect:    nil,
		},
		{
			name:      "error - token is revoked",
			revokedAt: vPtr(time.Now()),
			reason:    nil,
			expect:    model.ErrTokenAlreadyRevoked,
		},
		{
			name:      "error - empty reason",
			revokedAt: nil,
			reason:    vPtr(""),
			expect:    pkgerrs.ErrValueIsInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := model.RestoreRefreshSession(
				uuid.New(), uuid.New(), "hashed",
				time.Now(), time.Now().Add(time.Hour),
				tt.revokedAt, nil,
				nil, nil, nil)
			err := session.Revoke(tt.reason)
			if tt.expect == nil {
				require.NoError(t, err)

				var compareReason string
				if tt.reason != nil {
					compareReason = *tt.reason
				}

				assert.Equal(t, compareReason, *(session.RevokeReason()))
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expect)
				assert.Empty(t, session.RevokeReason())
			}
		})
	}
}
