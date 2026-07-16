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

func TestSendVerification_Success(t *testing.T) {
	app := setupE2E(t)
	ctx := context.Background()
	accountID, _, _ := app.createAccount(t, nil, nil, nil, nil, true)

	t.Run("Successfully sent", func(t *testing.T) {
		resp, err := app.client.SendVerification(ctx,
			&auth_v1.SendVerificationRequest{
				AccountId: accountID,
			},
		)
		require.NoError(t, err)
		require.True(t, resp.GetSent())
	})
}

func TestSendVerification_BadCases(t *testing.T) {
	app := setupE2E(t)
	ctx := context.Background()

	/*
		Prepare test data:
		1) Create another account and block it
		2) Create the third account and delete it
	*/
	_, blockedEmail, deletedEmail := gofakeit.Email(), gofakeit.Email(), gofakeit.Email()
	password := gofakeit.Password(true, true, true, true, true, 10)

	blockedAccID, _, _ := app.createAccount(t, &blockedEmail, &password, nil, nil, false)
	deletedAccID, _, _ := app.createAccount(t, &deletedEmail, &password, nil, nil, false)

	app.blockAccount(t, blockedAccID)
	app.deleteAccount(t, deletedAccID)

	type testCase struct {
		name          string
		accountID     string
		expectedCode  codes.Code
		expectedError string
	}

	var tests = []testCase{
		{
			name:          "Not Found - Account Doesn't Exist",
			accountID:     gofakeit.UUID(),
			expectedCode:  codes.NotFound,
			expectedError: "account not found",
		},
		{
			name:          "Failed Precondition - Account Is Blocked",
			accountID:     blockedAccID,
			expectedCode:  codes.FailedPrecondition,
			expectedError: "account is either blocked or not exists",
		},
		{
			name:          "Failed Precondition - Account Is Deleted",
			accountID:     deletedAccID,
			expectedCode:  codes.FailedPrecondition,
			expectedError: "account is either blocked or not exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := app.client.SendVerification(ctx,
				&auth_v1.SendVerificationRequest{AccountId: tt.accountID},
			)
			require.Error(t, err)
			assert.False(t, resp.GetSent())

			st, ok := status.FromError(err)
			require.True(t, ok, "expected a gRPC status error")
			assert.Equal(t, tt.expectedCode, st.Code())
			assert.Contains(t, st.Message(), tt.expectedError)
		})
	}
}
