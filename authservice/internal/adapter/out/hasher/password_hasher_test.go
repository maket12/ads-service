package hasher_test

import (
	"testing"

	"github.com/maket12/ads-service/authservice/internal/adapter/out/hasher"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordHasher_Hash(t *testing.T) {
	t.Parallel()

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
	var passwordHasher = hasher.NewBcryptHasher(hashCost)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := passwordHasher.Hash(tt.password)
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
	t.Parallel()

	const hashCost = 4
	var passwordHasher = hasher.NewBcryptHasher(hashCost)

	t.Run("success", func(t *testing.T) {
		var (
			password = "password-12345"
			hash, _  = passwordHasher.Hash(password)
		)
		result := passwordHasher.Compare(hash, password)
		assert.True(t, result)
	})

	t.Run("fail - not equal", func(t *testing.T) {
		var (
			password = "password-12345"
			hash     = "wrong-hash"
		)
		result := passwordHasher.Compare(hash, password)
		assert.False(t, result)
	})
}
