package model_test

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	pkgerrs "github.com/maket12/ads-service/authservice/pkg/errs"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testEmail    = "test@email.go"
	testPassword = "hashed-password"
)

func TestNewAccount(t *testing.T) {
	type testCase struct {
		name         string
		email        string
		passwordHash string
		expect       error
	}

	var tests = []testCase{
		{
			name:  "success",
			email: gofakeit.Email(),
			passwordHash: gofakeit.Password(
				true, false, false,
				false, false, 8,
			),
			expect: nil,
		},
		{
			name:   "empty email",
			email:  "",
			expect: pkgerrs.ErrValueIsRequired,
		},
		{
			name:         "empty password",
			email:        gofakeit.Email(),
			passwordHash: "",
			expect:       pkgerrs.ErrValueIsRequired,
		},
		{
			name:  "short email",
			email: "12@g",
			passwordHash: gofakeit.Password(
				true, false, false,
				false, false, 8,
			),
			expect: pkgerrs.ErrValueIsInvalid,
		},
		{
			name:  "invalid email",
			email: "123456789",
			passwordHash: gofakeit.Password(
				true, false, false,
				false, false, 8,
			),
			expect: pkgerrs.ErrValueIsInvalid,
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
				assert.Nil(t, acc.LastLoginAt())
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expect)
				assert.Nil(t, acc)
			}
		})
	}
}

func TestAccount_CanLogin(t *testing.T) {
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
				uuid.New(), testEmail, testPassword,
				tt.status, false, time.Now(),
				time.Now(), nil)
			assert.Equal(t, tt.expected, acc.CanLogin())
		})
	}
}

func TestAccount_IsBlocked(t *testing.T) {
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
				uuid.New(), testEmail, testPassword,
				tt.status, false, time.Now(),
				time.Now(), nil)
			assert.Equal(t, tt.expected, acc.IsBlocked())
		})
	}
}

func TestAccount_IsDeleted(t *testing.T) {
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
				uuid.New(), testEmail, testPassword,
				tt.status, false, time.Now(),
				time.Now(), nil)
			assert.Equal(t, tt.expected, acc.IsDeleted())
		})
	}
}

func TestAccount_Block(t *testing.T) {
	acc, _ := model.NewAccount(testEmail, testPassword)

	// First case: block successfully
	err := acc.Block()
	assert.NoError(t, err)
	assert.True(t, acc.IsBlocked())

	// Second case: failed to block
	err = acc.Block()
	assert.Error(t, err)
}

func TestAccount_Delete(t *testing.T) {
	acc, _ := model.NewAccount(testEmail, testPassword)

	// First case: delete successfully
	err := acc.Delete()
	assert.NoError(t, err)
	assert.True(t, acc.IsDeleted())

	// Second case: failed to delete
	err = acc.Delete()
	assert.Error(t, err)
}

func TestAccount_MarkLogin(t *testing.T) {
	acc, _ := model.NewAccount(testEmail, testPassword)

	// First case: login is succeeded
	err := acc.MarkLogin()
	assert.NoError(t, err)
	assert.NotNil(t, acc.LastLoginAt())

	// Second case: failed to log in (account is unreachable)
	_ = acc.Delete()
	err = acc.MarkLogin()
	assert.Error(t, err)
}

func TestAccount_VerifyEmail(t *testing.T) {
	acc, _ := model.NewAccount(testEmail, testPassword)
	acc.VerifyEmail()
	assert.True(t, acc.EmailVerified())
}
