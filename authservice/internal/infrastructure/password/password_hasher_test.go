package password_test

import (
	"testing"

	"github.com/maket12/meeting-rooms-api/internal/infrastructure/password"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordHasher_Hash(t *testing.T) {
	type testCase struct {
		name     string
		password string
		wantErr  bool
	}

	var tests = []testCase{
		{
			name:     "success",
			password: "password-12345",
			wantErr:  false,
		},
		{
			name:     "fail to hash",
			password: "it-is-a-really-long-password-to-check-what-will-happen-if-i-try-to-hash-it",
			wantErr:  true,
		},
	}

	const hashCost = 4
	var hasher = password.NewHasher(hashCost)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := hasher.Hash(tt.password)
			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, hash)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, hash)
			}
		})
	}
}

func TestPasswordHasher_Compare(t *testing.T) {
	const hashCost = 4
	var hasher = password.NewHasher(hashCost)

	t.Run("success", func(t *testing.T) {
		var (
			testPassword = "password-12345"
			hash, _      = hasher.Hash(testPassword)
		)
		result := hasher.Compare(hash, testPassword)
		assert.True(t, result)
	})

	t.Run("fail - not equal", func(t *testing.T) {
		var (
			testPassword = "password-12345"
			hash         = "wrong-hash"
		)
		result := hasher.Compare(hash, testPassword)
		assert.False(t, result)
	})
}
