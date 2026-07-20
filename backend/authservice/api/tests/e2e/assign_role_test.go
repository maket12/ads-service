//go:build e2e

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

func TestAssignRole_Success(t *testing.T) {
	app := setupE2E(t)
	ctx := context.Background()
	accountID, _, _ := app.createAccount(t, nil, nil, nil, nil, false)

	t.Run("Successfully assigned to admin", func(t *testing.T) {
		resp, err := app.client.AssignRole(ctx,
			&auth_v1.AssignRoleRequest{
				AccountId: accountID,
				Role:      "admin",
			},
		)
		require.NoError(t, err)
		require.True(t, resp.GetAssigned())
	})

	t.Run("Successfully assigned to user", func(t *testing.T) {
		resp, err := app.client.AssignRole(ctx,
			&auth_v1.AssignRoleRequest{
				AccountId: accountID,
				Role:      "user",
			},
		)
		require.NoError(t, err)
		require.True(t, resp.GetAssigned())
	})
}

func TestAssignRole_BadCases(t *testing.T) {
	app := setupE2E(t)
	ctx := context.Background()
	accountID, _, _ := app.createAccount(t, nil, nil, nil, nil, false)

	type testCase struct {
		name          string
		accountID     string
		role          string
		expectedCode  codes.Code
		expectedError string
	}

	var tests = []testCase{
		{
			name:          "Invalid Argument - Role Is Invalid",
			accountID:     accountID,
			role:          "hacker",
			expectedCode:  codes.InvalidArgument,
			expectedError: "account cannot be assigned to this role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := app.client.AssignRole(ctx, &auth_v1.AssignRoleRequest{
				AccountId: tt.accountID,
				Role:      tt.role,
			})
			require.Error(t, err)
			assert.False(t, resp.GetAssigned())

			st, ok := status.FromError(err)
			require.True(t, ok, "expected a gRPC status error")
			assert.Equal(t, tt.expectedCode, st.Code())
			assert.Contains(t, st.Message(), tt.expectedError)
		})
	}
}
