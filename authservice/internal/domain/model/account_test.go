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

func TestNewAccount(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name         string
		email        string
		passwordHash string
		expect       error
	}

	var tests = []testCase{
		{
			name:         "success",
			email:        "new-email@gmail.com",
			passwordHash: "new-password",
			expect:       nil,
		},
		{
			name:   "empty email",
			email:  "",
			expect: pkgerrs.ErrValueIsRequired,
		},
		{
			name:         "empty password",
			email:        "new-email@gmail.com",
			passwordHash: "",
			expect:       pkgerrs.ErrValueIsRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc, err := model.NewAccount(tt.email, tt.passwordHash)
			if tt.expect == nil {
				require.NoError(t, err)
				assert.NotNil(t, acc.ID())
				assert.Equal(t, tt.email, acc.Email())
				assert.Equal(t, tt.passwordHash, acc.PasswordHash())
				assert.Equal(t, acc.Status(), model.AccountActive)
				assert.False(t, acc.EmailVerified())
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, pkgerrs.ErrValueIsRequired)
				assert.Nil(t, acc)
			}
		})
	}
}

func TestAccount_CanLogin(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name     string
		status   model.AccountStatus
		expected bool
	}

	var tests = []testCase{
		{
			name:     "active - can login",
			status:   model.AccountActive,
			expected: true,
		},
		{
			name:     "blocked - cannot login",
			status:   model.AccountBlocked,
			expected: false,
		},
		{
			name:     "deleted - cannot login",
			status:   model.AccountDeleted,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := model.RestoreAccount(
				uuid.New(), "email.com", "password",
				tt.status, false, time.Now(),
				time.Now(), nil)
			assert.Equal(t, tt.expected, acc.CanLogin())
		})
	}
}

func TestAccount_IsBlocked(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name     string
		status   model.AccountStatus
		expected bool
	}

	var tests = []testCase{
		{
			name:     "active - false",
			status:   model.AccountActive,
			expected: false,
		},
		{
			name:     "blocked - true",
			status:   model.AccountBlocked,
			expected: true,
		},
		{
			name:     "deleted - false",
			status:   model.AccountDeleted,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := model.RestoreAccount(
				uuid.New(), "email.com", "password",
				tt.status, false, time.Now(),
				time.Now(), nil)
			assert.Equal(t, tt.expected, acc.IsBlocked())
		})
	}
}

func TestAccount_IsDeleted(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name     string
		status   model.AccountStatus
		expected bool
	}

	var tests = []testCase{
		{
			name:     "active - false",
			status:   model.AccountActive,
			expected: false,
		},
		{
			name:     "blocked - false",
			status:   model.AccountBlocked,
			expected: false,
		},
		{
			name:     "deleted - true",
			status:   model.AccountDeleted,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := model.RestoreAccount(
				uuid.New(), "email.com", "password",
				tt.status, false, time.Now(),
				time.Now(), nil)
			assert.Equal(t, tt.expected, acc.IsDeleted())
		})
	}
}

func TestAccount_StatedChanges(t *testing.T) {
	t.Parallel()

	// Test account
	acc, _ := model.NewAccount("test@email.go", "password")
	initialUpdatedAt := acc.UpdatedAt()

	// Wait a millisecond to change updatedAt
	time.Sleep(time.Millisecond)

	t.Run("verify email", func(t *testing.T) {
		acc.VerifyEmail()
		assert.True(t, acc.EmailVerified())
		assert.True(t, acc.UpdatedAt().After(initialUpdatedAt))
	})
	t.Run("mark login", func(t *testing.T) {
		acc.MarkLogin()
		assert.True(t, acc.LastLoginAt().After(initialUpdatedAt))
		assert.True(t, acc.UpdatedAt().After(initialUpdatedAt))
	})
	t.Run("block account", func(t *testing.T) {
		acc.Block()
		assert.Equal(t, model.AccountBlocked, acc.Status())
		assert.True(t, acc.UpdatedAt().After(initialUpdatedAt))
	})
	t.Run("delete account", func(t *testing.T) {
		acc.Delete()
		assert.Equal(t, model.AccountDeleted, acc.Status())
		assert.True(t, acc.UpdatedAt().After(initialUpdatedAt))
	})
}
