///go:build e2e

package e2e

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/maket12/ads-service/backend/adservice/v2/pkg/generated/ad_v1"
	"github.com/maket12/ads-service/backend/adservice/v2/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRejectAd_Success(t *testing.T) {
	app := setupE2E(t)

	adID, _ := app.createAd(t, nil, nil)
	ctx := utils.PackAccountIDForGRPC(context.Background(), gofakeit.UUID())
	ctx = utils.PackAccountRoleForGRPC(ctx, "admin")

	resp, err := app.client.RejectAd(ctx, &ad_v1.RejectAdRequest{Id: adID})

	require.NoError(t, err)
	require.True(t, resp.GetSuccess())
}

func TestRejectAd_BadCases(t *testing.T) {
	app := setupE2E(t)

	adminID := gofakeit.UUID()
	adID, sellerID := app.createAd(t, nil, nil)
	rejectedAdID, _ := app.createAd(t, &sellerID, nil)
	app.rejectAd(t, rejectedAdID, adminID)

	type testCase struct {
		name          string
		adminID       string
		adID          string
		expectedCode  codes.Code
		expectedError string
	}

	var tests = []testCase{
		{
			name:          "Not Found - Ad Doesn't Exist",
			adminID:       adminID,
			adID:          gofakeit.UUID(),
			expectedCode:  codes.NotFound,
			expectedError: "ad not found",
		},
		{
			name:          "Failed Precondition - Ad Has Been Rejected",
			adminID:       adminID,
			adID:          rejectedAdID,
			expectedCode:  codes.FailedPrecondition,
			expectedError: "ad has been already published or not available",
		},
		{
			name:          "Unauthenticated",
			adminID:       "",
			adID:          adID,
			expectedCode:  codes.Unauthenticated,
			expectedError: "you must be authenticated to make this request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := utils.PackAccountIDForGRPC(context.Background(), tt.adminID)
			ctx = utils.PackAccountRoleForGRPC(ctx, "admin")

			resp, err := app.client.RejectAd(ctx, &ad_v1.RejectAdRequest{
				Id: tt.adID,
			})

			require.Error(t, err)
			assert.False(t, resp.GetSuccess())

			st, ok := status.FromError(err)
			require.True(t, ok, "expected a gRPC status error")
			assert.Equal(t, tt.expectedCode, st.Code())
			assert.Contains(t, st.Message(), tt.expectedError)
		})
	}
}
