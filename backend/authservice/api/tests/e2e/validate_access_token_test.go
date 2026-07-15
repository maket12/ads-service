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

func TestValidateAccessToken_Success(t *testing.T) {
	app := setupE2E(t)
	ctx := context.Background()

	// Get account id and token
	accountID, access, _ := app.createAccount(t, nil, nil, nil, nil, true)

	resp, err := app.client.ValidateAccessToken(ctx, &auth_v1.ValidateAccessTokenRequest{
		AccessToken: access,
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.GetAccountId())
	require.Equal(t, accountID, resp.GetAccountId())
	require.NotEmpty(t, resp.GetRole())
}

func TestValidateAccessToken_BadCases(t *testing.T) {
	app := setupE2E(t)
	ctx := context.Background()
	
	type testCase struct {
		name          string
		accessToken   string
		expectedCode  codes.Code
		expectedError string
	}

	var tests = []testCase{
		{
			name:          "Invalid Argument - Not a Token",
			accessToken:   "not-a-token",
			expectedCode:  codes.InvalidArgument,
			expectedError: "access token is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := app.client.ValidateAccessToken(ctx, &auth_v1.ValidateAccessTokenRequest{
				AccessToken: tt.accessToken,
			})

			require.Error(t, err)
			assert.Empty(t, resp.GetAccountId())
			assert.Empty(t, resp.GetRole())

			st, ok := status.FromError(err)
			require.True(t, ok, "expected a gRPC status error")
			assert.Equal(t, tt.expectedCode, st.Code())
			assert.Contains(t, st.Message(), tt.expectedError)
		})
	}
}
