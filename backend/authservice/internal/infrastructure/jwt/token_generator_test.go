package jwt_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenGenerator_Success(t *testing.T) {
	var (
		secret     = "secret-key"
		ttl        = time.Hour
		testUserID = uuid.New()
		testRole   = "admin"
	)

	gen := NewGenerator(secret, ttl)

	// Generate a token
	token, err := gen.Generate(testUserID, testRole)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate this token
	userID, role, err := gen.Validate(token)
	require.NoError(t, err)
	assert.Equal(t, testUserID, userID)
	assert.Equal(t, testRole, role)
}

func TestTokenGenerator_Expired(t *testing.T) {
	var (
		secret = "secret-key"
		ttl    = time.Millisecond // Set the too short time
	)

	gen := NewGenerator(secret, ttl)

	// Generate a token
	token, _ := gen.Generate(uuid.New(), "admin")

	// Wait some time to ensure leeway checking will throw false
	time.Sleep(3 * time.Second)

	// Validate an expired token
	userID, role, err := gen.Validate(token)
	require.Error(t, err)
	assert.Empty(t, userID)
	assert.Empty(t, role)
}

func TestTokenGenerator_InvalidSecret(t *testing.T) {
	var (
		validSecret   = "valid-key"
		invalidSecret = "no-valid-key"
		ttl           = time.Hour
	)

	genValid := NewGenerator(validSecret, ttl)
	genInvalid := NewGenerator(invalidSecret, ttl)

	// Create a token using the first generator
	token, _ := genValid.Generate(uuid.New(), "user")

	// Validate it using the second generator
	uID, role, err := genInvalid.Validate(token)
	require.Error(t, err)
	assert.Empty(t, uID)
	assert.Empty(t, role)
}

func TestTokenGenerator_RandomString(t *testing.T) {
	var (
		secret = "valid-key"
		ttl    = time.Hour
	)

	gen := NewGenerator(secret, ttl)

	// Validate a random string (not a token)
	uID, role, err := gen.Validate("not-a-token")
	require.Error(t, err)
	assert.Empty(t, uID)
	assert.Empty(t, role)
}
