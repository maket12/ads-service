///go:build e2e

package e2e

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/maket12/ads-service/authservice/pkg/generated/auth_v1"
	"github.com/maket12/ads-service/authservice/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRefreshSession_Success(t *testing.T) {
	app := setupE2E(t)
	ctx := context.Background()

	email := gofakeit.Email()
	password := gofakeit.Password(true, true, true, true, true, 10)
	ip := gofakeit.IPv4Address()
	ua := gofakeit.UserAgent()

	_, _, oldRefreshToken := app.createAccount(t,
		&email, &password,
		&ip, &ua, true,
	)

	resp, err := app.client.RefreshSession(ctx, &auth_v1.RefreshSessionRequest{
		OldRefreshToken: oldRefreshToken,
		Ip:              utils.VPtr(ip),
		UserAgent:       utils.VPtr(ua),
	})

	require.NoError(t, err)
	require.NotEmpty(t, resp.GetAccessToken())
	require.NotEmpty(t, resp.GetRefreshToken())
	require.NotEqual(t, oldRefreshToken, resp.GetRefreshToken())

	// The old token must now be rotated out — using it again should fail.
	_, err = app.client.RefreshSession(ctx, &auth_v1.RefreshSessionRequest{
		OldRefreshToken: oldRefreshToken,
		Ip:              utils.VPtr(ip),
		UserAgent:       utils.VPtr(ua),
	})
	require.Error(t, err)
}

func TestRefreshSession_BadCases(t *testing.T) {
	app := setupE2E(t)
	ctx := context.Background()

	type testCase struct {
		name          string
		setup         func(t *testing.T) (token, ip, ua string)
		expectedCode  codes.Code
		expectedError string
	}

	var tests = []testCase{
		{
			name: "Invalid Argument - Garbage Token",
			setup: func(t *testing.T) (string, string, string) {
				return "not-a-valid-jwt", gofakeit.IPv4Address(), gofakeit.UserAgent()
			},
			expectedCode:  codes.InvalidArgument,
			expectedError: "refresh token is invalid or not found",
		},
		{
			name: "Unauthenticated - Reused/Rotated Token (Compromised Reuse)",
			setup: func(t *testing.T) (string, string, string) {
				email := gofakeit.Email()
				password := gofakeit.Password(true, true, true, true, true, 10)
				ip := gofakeit.IPv4Address()
				ua := gofakeit.UserAgent()

				_, _, oldToken := app.createAccount(t, &email, &password, &ip, &ua, true)

				// Rotate once — this revokes oldToken and creates a new session.
				refreshResp, err := app.client.RefreshSession(ctx, &auth_v1.RefreshSessionRequest{
					OldRefreshToken: oldToken,
					Ip:              utils.VPtr(ip),
					UserAgent:       utils.VPtr(ua),
				})
				require.NoError(t, err)
				require.NotEmpty(t, refreshResp.GetRefreshToken())

				// Reusing the already-rotated old token triggers breach detection.
				return oldToken, ip, ua
			},
			expectedCode:  codes.InvalidArgument,
			expectedError: "refresh token is invalid or not found",
		},
		{
			name: "Unauthenticated - Suspicious Environment (IP/UA Mismatch)",
			setup: func(t *testing.T) (string, string, string) {
				email := gofakeit.Email()
				password := gofakeit.Password(true, true, true, true, true, 10)
				originalIP := gofakeit.IPv4Address()
				originalUA := gofakeit.UserAgent()

				_, _, token := app.createAccount(t, &email, &password, &originalIP, &originalUA, true)

				// Attempt refresh from a different IP/UA than the session was created with.
				return token, gofakeit.IPv4Address(), gofakeit.UserAgent()
			},
			expectedCode:  codes.InvalidArgument,
			expectedError: "refresh token is invalid or not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, ip, ua := tt.setup(t)

			resp, err := app.client.RefreshSession(ctx, &auth_v1.RefreshSessionRequest{
				OldRefreshToken: token,
				Ip:              utils.VPtr(ip),
				UserAgent:       utils.VPtr(ua),
			})

			require.Error(t, err)
			assert.Empty(t, resp.GetAccessToken())
			assert.Empty(t, resp.GetRefreshToken())

			st, ok := status.FromError(err)
			require.True(t, ok, "expected a gRPC status error")
			assert.Equal(t, tt.expectedCode, st.Code())
			assert.Contains(t, st.Message(), tt.expectedError)
		})
	}
}

// TestRefreshSession_CompromisedReuse_RevokesDescendants verifies that once a
// rotated (already-used) refresh token is replayed, the breach-detection logic
// revokes the entire session chain — including the newer, otherwise-valid
// descendant token issued by the legitimate rotation.
func TestRefreshSession_CompromisedReuse_RevokesDescendants(t *testing.T) {
	app := setupE2E(t)
	ctx := context.Background()

	email := gofakeit.Email()
	password := gofakeit.Password(true, true, true, true, true, 10)
	ip := gofakeit.IPv4Address()
	ua := gofakeit.UserAgent()

	_, _, oldToken := app.createAccount(t, &email, &password, &ip, &ua, true)

	// Legitimate rotation.
	refreshResp, err := app.client.RefreshSession(ctx, &auth_v1.RefreshSessionRequest{
		OldRefreshToken: oldToken,
		Ip:              utils.VPtr(ip),
		UserAgent:       utils.VPtr(ua),
	})
	require.NoError(t, err)
	newToken := refreshResp.GetRefreshToken()
	require.NotEmpty(t, newToken)

	// Attacker replays the old (already-rotated) token — should be rejected
	// and should revoke the descendant chain as a side effect.
	_, err = app.client.RefreshSession(ctx, &auth_v1.RefreshSessionRequest{
		OldRefreshToken: oldToken,
		Ip:              utils.VPtr(ip),
		UserAgent:       utils.VPtr(ua),
	})
	require.Error(t, err)

	// The legitimate new token, issued by the first rotation, must now be
	// unusable too, since its whole chain was revoked as compromised.
	_, err = app.client.RefreshSession(ctx, &auth_v1.RefreshSessionRequest{
		OldRefreshToken: newToken,
		Ip:              utils.VPtr(ip),
		UserAgent:       utils.VPtr(ua),
	})
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok, "expected a gRPC status error")
	assert.Equal(t, codes.InvalidArgument, st.Code())
}
