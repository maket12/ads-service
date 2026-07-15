///go:build e2e

package e2e

import (
	"context"
	"testing"

	"github.com/maket12/ads-service/authservice/pkg/generated/auth_v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLogout_Success(t *testing.T) {
	app := setupE2E(t)
	ctx := context.Background()

	// Get refresh token (login)
	_, _, refreshToken := app.createAccount(t, nil, nil, nil, nil, true)

	resp, err := app.client.Logout(ctx, &auth_v1.LogoutRequest{
		RefreshToken: refreshToken,
	})
	require.NoError(t, err)
	require.True(t, resp.GetLogout())
}

func TestLogout_BadCases(t *testing.T) {
	app := setupE2E(t)
	ctx := context.Background()

	// Get a refresh token and then revoke it for checking the special case
	_, _, revokedToken := app.createAccount(t, nil, nil, nil, nil, true)
	app.logout(t, revokedToken)

	type testCase struct {
		name          string
		refreshToken  string
		expectedCode  codes.Code
		expectedError string
	}

	var tests = []testCase{
		{
			name:          "Invalid Argument - Not a Token",
			refreshToken:  "not-a-token",
			expectedCode:  codes.InvalidArgument,
			expectedError: "refresh token is invalid or not found",
		},
		{
			name:          "Failed Precondition - Session Already Revoked",
			refreshToken:  revokedToken,
			expectedCode:  codes.FailedPrecondition,
			expectedError: "session is already expired or revoked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := app.client.Logout(ctx, &auth_v1.LogoutRequest{
				RefreshToken: tt.refreshToken,
			})

			require.Error(t, err)
			assert.False(t, resp.GetLogout())

			st, ok := status.FromError(err)
			require.True(t, ok, "expected a gRPC status error")
			assert.Equal(t, tt.expectedCode, st.Code())
			assert.Contains(t, st.Message(), tt.expectedError)
		})
	}
}
