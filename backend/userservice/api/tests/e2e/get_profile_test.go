///go:build e2e

package e2e

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/maket12/ads-service/userservice/pkg/generated/user_v1"
	"github.com/maket12/ads-service/userservice/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetProfile_Success(t *testing.T) {
	app := setupE2E(t)
	accountID := app.createProfile(t, nil)
	ctx := utils.PackAccountIDForGRPC(context.Background(), accountID)

	resp, err := app.client.GetProfile(ctx, &user_v1.GetProfileRequest{})

	require.NoError(t, err)
	require.Equal(t, accountID, resp.AccountId)
}

func TestGetProfile_BadCases(t *testing.T) {
	app := setupE2E(t)

	type testCase struct {
		name          string
		accountID     string
		expectedCode  codes.Code
		expectedError string
	}

	var tests = []testCase{
		{
			name:          "Not Found - Profile Doesn't Exist",
			accountID:     uuid.New().String(),
			expectedCode:  codes.NotFound,
			expectedError: "profile not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := utils.PackAccountIDForGRPC(context.Background(), tt.accountID)
			resp, err := app.client.GetProfile(ctx, &user_v1.GetProfileRequest{})

			require.Error(t, err)
			assert.Empty(t, resp.GetAccountId())
			assert.Empty(t, resp.GetFirstName())
			assert.Empty(t, resp.GetLastName())
			assert.Empty(t, resp.GetPhone())
			assert.Empty(t, resp.GetAvatarUrl())
			assert.Empty(t, resp.GetBio())
			assert.Empty(t, resp.GetUpdatedAt())

			st, ok := status.FromError(err)
			require.True(t, ok, "expected a gRPC status error")
			assert.Equal(t, tt.expectedCode, st.Code())
			assert.Contains(t, st.Message(), tt.expectedError)
		})
	}
}
