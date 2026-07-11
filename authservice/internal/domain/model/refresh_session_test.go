package model_test

import (
	"testing"
	"time"

	"github.com/maket12/ads-service/authservice/internal/domain/model"
	pkgerrs "github.com/maket12/ads-service/authservice/pkg/errs"
	"github.com/maket12/ads-service/authservice/pkg/utils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTokenHash = "hashed-token"

func TestNewRefreshSession(t *testing.T) {
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
			tokenHash: testTokenHash,
			ttl:       time.Minute,
			expect:    nil,
		},
		{
			name:      "nullable session id",
			id:        uuid.Nil,
			accountID: uuid.New(),
			tokenHash: testTokenHash,
			ttl:       time.Minute,
			expect:    pkgerrs.ErrValueIsInvalid,
		},
		{
			name:      "nullable account id",
			id:        uuid.New(),
			accountID: uuid.Nil,
			tokenHash: testTokenHash,
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
			tokenHash:   testTokenHash,
			rotatedFrom: utils.VPtr(uuid.Nil),
			ttl:         time.Minute,
			expect:      pkgerrs.ErrValueIsInvalid,
		},
		{
			name:      "invalid ttl (not positive)",
			id:        uuid.New(),
			accountID: uuid.New(),
			tokenHash: testTokenHash,
			ip:        utils.VPtr("123.021.234.0"),
			userAgent: utils.VPtr("Mozilla/5.0"),
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
				uuid.New(), uuid.New(), testTokenHash,
				time.Now(), tt.expiresAt, nil, nil,
				nil, nil, nil)
			assert.Equal(t, tt.expect, session.IsExpired())
		})
	}
}

func TestRefreshSession_IsRevoked(t *testing.T) {
	type testCase struct {
		name      string
		revokedAt *time.Time
		expect    bool
	}

	var tests = []testCase{
		{
			name:      "session is revoked - true",
			revokedAt: utils.VPtr(time.Now()),
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
				uuid.New(), uuid.New(), testTokenHash,
				time.Now(), time.Now().Add(time.Hour),
				tt.revokedAt, nil,
				nil, nil, nil)
			assert.Equal(t, tt.expect, session.IsRevoked())
		})
	}
}

func TestRefreshSession_IsActive(t *testing.T) {
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
			revokedAt: utils.VPtr(time.Now().Add(time.Minute * -1)),
			expect:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := model.RestoreRefreshSession(
				uuid.New(), uuid.New(), testTokenHash,
				time.Now(), tt.expiresAt, tt.revokedAt, nil,
				nil, nil, nil)
			assert.Equal(t, tt.expect, session.IsActive())
		})
	}
}

func TestRefreshSession_RevokeByLogout(t *testing.T) {
	session, _ := model.NewRefreshSession(
		uuid.New(), uuid.New(),
		testTokenHash, nil,
		nil, nil, time.Minute,
	)

	// First case - revoke successfully
	err := session.RevokeByLogout()
	require.NoError(t, err)
	assert.True(t, session.IsRevoked())
	assert.Equal(t, model.ReasonLogout, *session.RevokeReason())

	// Second case - revoke failed
	err = session.RevokeByLogout()
	require.Error(t, err)
	assert.ErrorIs(t, err, model.ErrCannotRevokeToken)
}

func TestRefreshSession_RevokeByRotation(t *testing.T) {
	session, _ := model.NewRefreshSession(
		uuid.New(), uuid.New(),
		testTokenHash, nil,
		nil, nil, time.Minute,
	)

	// First case - revoke successfully
	err := session.RevokeByRotation()
	require.NoError(t, err)
	assert.True(t, session.IsRevoked())
	assert.Equal(t, model.ReasonTokenRotation, *session.RevokeReason())

	// Second case - revoke failed
	err = session.RevokeByRotation()
	require.Error(t, err)
	assert.ErrorIs(t, err, model.ErrCannotRevokeToken)
}

func TestRefreshSession_RevokeBySuspiciousEnv(t *testing.T) {
	session, _ := model.NewRefreshSession(
		uuid.New(), uuid.New(),
		testTokenHash, nil,
		nil, nil, time.Minute,
	)

	// First case - revoke successfully
	err := session.RevokeBySuspiciousEnv()
	require.NoError(t, err)
	assert.True(t, session.IsRevoked())
	assert.Equal(t, model.ReasonSuspiciousEnv, *session.RevokeReason())

	// Second case - revoke failed
	err = session.RevokeBySuspiciousEnv()
	require.Error(t, err)
	assert.ErrorIs(t, err, model.ErrCannotRevokeToken)
}
