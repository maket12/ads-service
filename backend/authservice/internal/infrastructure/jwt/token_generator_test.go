package jwt_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/maket12/ads-service/authservice/internal/domain/port"
	infrajwt "github.com/maket12/ads-service/authservice/internal/infrastructure/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAccessSecret  = "access-secret-key"
	testRefreshSecret = "refresh-secret-key"
	testAccessTTL     = time.Hour
	testRefreshTTL    = 24 * time.Hour
)

func newTestGenerator() *infrajwt.Generator {
	return infrajwt.NewGenerator(
		testAccessSecret, testRefreshSecret,
		testAccessTTL, testRefreshTTL,
	)
}

func TestGeneratePair_Success(t *testing.T) {
	gen := newTestGenerator()
	ctx := context.Background()

	var (
		accountID = uuid.New()
		role      = "admin"
		sessionID = uuid.New()
	)

	pair, err := gen.GeneratePair(ctx, accountID, role, sessionID)

	require.NoError(t, err)
	require.NotNil(t, pair)
	assert.NotEmpty(t, pair.Access)
	assert.NotEmpty(t, pair.Refresh)
	assert.NotEqual(t, pair.Access, pair.Refresh)
}

func TestValidateAccessToken_Success(t *testing.T) {
	gen := newTestGenerator()
	ctx := context.Background()

	var (
		accountID = uuid.New()
		role      = "admin"
		sessionID = uuid.New()
	)

	pair, err := gen.GeneratePair(ctx, accountID, role, sessionID)
	require.NoError(t, err)

	gotAccountID, gotRole, err := gen.ValidateAccessToken(ctx, pair.Access)

	require.NoError(t, err)
	assert.Equal(t, accountID, gotAccountID)
	assert.Equal(t, role, gotRole)
}

func TestValidateAccessToken_WrongSecret(t *testing.T) {
	genValid := newTestGenerator()
	genInvalid := infrajwt.NewGenerator(
		"another-access-secret", testRefreshSecret,
		testAccessTTL, testRefreshTTL,
	)
	ctx := context.Background()

	pair, err := genValid.GeneratePair(ctx, uuid.New(), "admin", uuid.New())
	require.NoError(t, err)

	accountID, role, err := genInvalid.ValidateAccessToken(ctx, pair.Access)

	require.Error(t, err)
	assert.Equal(t, uuid.Nil, accountID)
	assert.Empty(t, role)
}

func TestValidateAccessToken_RandomString(t *testing.T) {
	gen := newTestGenerator()
	ctx := context.Background()

	accountID, role, err := gen.ValidateAccessToken(ctx, "not-a-token")

	require.Error(t, err)
	assert.Equal(t, uuid.Nil, accountID)
	assert.Empty(t, role)
}

func TestValidateAccessToken_WrongTokenType(t *testing.T) {
	gen := newTestGenerator()
	ctx := context.Background()

	// A refresh token passed where an access token is expected
	pair, err := gen.GeneratePair(ctx, uuid.New(), "admin", uuid.New())
	require.NoError(t, err)

	accountID, role, err := gen.ValidateAccessToken(ctx, pair.Refresh)

	require.Error(t, err)
	assert.Equal(t, uuid.Nil, accountID)
	assert.Empty(t, role)
}

func TestValidateRefreshToken_Success(t *testing.T) {
	gen := newTestGenerator()
	ctx := context.Background()

	var (
		accountID = uuid.New()
		sessionID = uuid.New()
	)

	pair, err := gen.GeneratePair(ctx, accountID, "admin", sessionID)
	require.NoError(t, err)

	gotAccountID, gotSessionID, err := gen.ValidateRefreshToken(ctx, pair.Refresh)

	require.NoError(t, err)
	assert.Equal(t, accountID, gotAccountID)
	assert.Equal(t, sessionID, gotSessionID)
}

func TestValidateRefreshToken_WrongSecret(t *testing.T) {
	genValid := newTestGenerator()
	genInvalid := infrajwt.NewGenerator(
		testAccessSecret, "another-refresh-secret",
		testAccessTTL, testRefreshTTL,
	)
	ctx := context.Background()

	pair, err := genValid.GeneratePair(ctx, uuid.New(), "admin", uuid.New())
	require.NoError(t, err)

	accountID, sessionID, err := genInvalid.ValidateRefreshToken(ctx, pair.Refresh)

	require.Error(t, err)
	assert.False(t, errors.Is(err, port.ErrTokenExpired))
	assert.Equal(t, uuid.Nil, accountID)
	assert.Equal(t, uuid.Nil, sessionID)
}

func TestValidateRefreshToken_RandomString(t *testing.T) {
	gen := newTestGenerator()
	ctx := context.Background()

	accountID, sessionID, err := gen.ValidateRefreshToken(ctx, "not-a-token")

	require.Error(t, err)
	assert.Equal(t, uuid.Nil, accountID)
	assert.Equal(t, uuid.Nil, sessionID)
}

func TestValidateRefreshToken_WrongTokenType(t *testing.T) {
	gen := newTestGenerator()
	ctx := context.Background()

	// An access token passed where a refresh token is expected
	pair, err := gen.GeneratePair(ctx, uuid.New(), "admin", uuid.New())
	require.NoError(t, err)

	accountID, sessionID, err := gen.ValidateRefreshToken(ctx, pair.Access)

	require.Error(t, err)
	assert.Equal(t, uuid.Nil, accountID)
	assert.Equal(t, uuid.Nil, sessionID)
}

func TestGeneratePair_TokenClaims(t *testing.T) {
	gen := newTestGenerator()
	ctx := context.Background()

	var (
		accountID = uuid.New()
		role      = "moderator"
		sessionID = uuid.New()
	)

	pair, err := gen.GeneratePair(ctx, accountID, role, sessionID)
	require.NoError(t, err)

	// Parse access token claims without validation, to check raw structure
	accessParsed, _, err := new(jwt.Parser).ParseUnverified(pair.Access, jwt.MapClaims{})
	require.NoError(t, err)
	accessClaims, ok := accessParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)
	assert.Equal(t, accountID.String(), accessClaims["sub"])
	assert.Equal(t, role, accessClaims["role"])
	assert.Equal(t, port.AccessToken.String(), accessClaims["type"])

	// Parse refresh token claims without validation
	refreshParsed, _, err := new(jwt.Parser).ParseUnverified(pair.Refresh, jwt.MapClaims{})
	require.NoError(t, err)
	refreshClaims, ok := refreshParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)
	assert.Equal(t, accountID.String(), refreshClaims["sub"])
	assert.Equal(t, sessionID.String(), refreshClaims["jti"])
	assert.Equal(t, port.RefreshToken.String(), refreshClaims["type"])
}
