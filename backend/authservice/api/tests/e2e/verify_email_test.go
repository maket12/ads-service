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

func TestVerifyEmail_Success(t *testing.T) {
	app := setupE2E(t)
	ctx := context.Background()

	email := gofakeit.Email()
	accountID, _, _ := app.createAccount(t, utils.VPtr(email), nil, nil, nil, false)
	token := app.sendToken(t, accountID, email, false)

	t.Run("Successfully verified", func(t *testing.T) {
		resp, err := app.client.VerifyEmail(ctx,
			&auth_v1.VerifyEmailRequest{
				Token: token,
			},
		)
		require.NoError(t, err)
		require.True(t, resp.GetVerified())
	})
}

func TestVerifyEmail_BadCases(t *testing.T) {
	app := setupE2E(t)
	ctx := context.Background()

	email := gofakeit.Email()
	accountID, _, _ := app.createAccount(t, utils.VPtr(email), nil, nil, nil, false)
	expiredToken := app.sendToken(t, accountID, email, true)

	type testCase struct {
		name          string
		token         string
		expectedCode  codes.Code
		expectedError string
	}

	var tests = []testCase{
		{
			name:          "Failed Precondition - Token Is Expired",
			token:         expiredToken,
			expectedCode:  codes.FailedPrecondition,
			expectedError: "verification token has been expired",
		},
		{
			name:          "Not Found - Random Token",
			token:         gofakeit.UUID(),
			expectedCode:  codes.NotFound,
			expectedError: "verification token not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := app.client.VerifyEmail(ctx,
				&auth_v1.VerifyEmailRequest{Token: tt.token},
			)
			require.Error(t, err)
			assert.False(t, resp.GetVerified())

			st, ok := status.FromError(err)
			require.True(t, ok, "expected a gRPC status error")
			assert.Equal(t, tt.expectedCode, st.Code())
			assert.Contains(t, st.Message(), tt.expectedError)
		})
	}
}
