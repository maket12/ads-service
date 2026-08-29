//go:build e2e

package e2e

import (
	"testing"

	"github.com/google/uuid"
	"github.com/maket12/ads-service/backend/authservice/api/proto/generated/auth_v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestBlockAccount_Success(t *testing.T) {
	app := setupE2E(t)
	accountID, _, _ := app.createAccount(t, nil, nil, nil, nil, false)

	resp, err := app.client.BlockAccount(app.adminCtx(),
		&auth_v1.BlockAccountRequest{
			AccountId: accountID,
		},
	)
	require.NoError(t, err)
	require.True(t, resp.Blocked)
}

func TestBlockAccount_BadCases(t *testing.T) {
	app := setupE2E(t)

	blockedAccID, _, _ := app.createAccount(t, nil, nil, nil, nil, false)
	app.blockAccount(t, blockedAccID)

	deletedAccID, _, _ := app.createAccount(t, nil, nil, nil, nil, false)
	app.deleteAccount(t, deletedAccID)

	adminAccID, _, _ := app.createAccount(t, nil, nil, nil, nil, false)
	app.assignToAdmin(t, adminAccID)

	type testCase struct {
		name          string
		accountID     string
		expectedCode  codes.Code
		expectedError string
	}

	var tests = []testCase{
		{
			name:          "Invalid Argument - Account ID",
			accountID:     uuid.Nil.String(),
			expectedCode:  codes.InvalidArgument,
			expectedError: "invalid account id",
		},
		{
			name:          "Not Found - Account Doesn't Exist",
			accountID:     uuid.NewString(),
			expectedCode:  codes.NotFound,
			expectedError: "account not found",
		},
		{
			name:          "Failed Precondition - Insufficient Permissions",
			accountID:     adminAccID,
			expectedCode:  codes.FailedPrecondition,
			expectedError: "account cannot be blocked",
		},
		{
			name:          "Failed Precondition - Account Is Already Blocked",
			accountID:     blockedAccID,
			expectedCode:  codes.FailedPrecondition,
			expectedError: "account cannot be blocked",
		},
		{
			name:          "Failed Precondition - Account Is Deleted",
			accountID:     deletedAccID,
			expectedCode:  codes.FailedPrecondition,
			expectedError: "account cannot be blocked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := app.client.BlockAccount(app.adminCtx(), &auth_v1.BlockAccountRequest{
				AccountId: tt.accountID,
			})
			require.Error(t, err)
			assert.False(t, resp.GetBlocked())

			st, ok := status.FromError(err)
			require.True(t, ok, "expected a gRPC status error")
			assert.Equal(t, tt.expectedCode, st.Code())
			assert.Contains(t, st.Message(), tt.expectedError)
		})
	}
}
