///go:build e2e

package e2e

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
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

	/*
		Prepare test data:
		1) Create an account and block it
		2) Create another account and delete it
	*/
	blockedEmail, deletedEmail := gofakeit.Email(), gofakeit.Email()
	password := gofakeit.Password(true, true, true, true, true, 10)

	blockedAccID, blockedAccess, _ := app.createAccount(t, &blockedEmail, &password, nil, nil, true)
	deletedAccID, deletedAccess, _ := app.createAccount(t, &deletedEmail, &password, nil, nil, true)

	app.blockAccount(t, blockedAccID)
	app.deleteAccount(t, deletedAccID)

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
		{
			name:          "Failed Precondition - Account Is Blocked",
			accessToken:   blockedAccess,
			expectedCode:  codes.FailedPrecondition,
			expectedError: "account is either blocked or not exists",
		},
		{
			name:          "Failed Precondition - Account Is Deleted",
			accessToken:   deletedAccess,
			expectedCode:  codes.FailedPrecondition,
			expectedError: "account is either blocked or not exists",
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
