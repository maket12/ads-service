//go:build e2e

package e2e

import (
	"context"
	"testing"

	"github.com/maket12/ads-service/backend/adservice/pkg/generated/ad_v1"
	"github.com/maket12/ads-service/backend/authservice/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestListAds_Success(t *testing.T) {
	app := setupE2E(t)

	fstAdID, sellerID := app.createAd(t, nil, nil)
	sndAdID, _ := app.createAd(t, &sellerID, nil)
	ctx := utils.PackAccountIDForGRPC(context.Background(), sellerID)

	resp, err := app.client.ListAds(ctx, &ad_v1.ListAdsRequest{})
	require.NoError(t, err)
	require.Len(t, resp.GetAds(), 2)
	require.Equal(t, fstAdID, resp.GetAds()[0].GetId())
	require.Equal(t, sndAdID, resp.GetAds()[1].GetId())
}

func TestListAds_BadCases(t *testing.T) {
	app := setupE2E(t)

	type testCase struct {
		name          string
		sellerID      string
		expectedCode  codes.Code
		expectedError string
	}

	var tests = []testCase{
		{
			name:          "Unauthenticated - missing seller id",
			sellerID:      "",
			expectedCode:  codes.Unauthenticated,
			expectedError: "you must be authenticated to make this request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := utils.PackAccountIDForGRPC(context.Background(), tt.sellerID)
			resp, err := app.client.ListAds(ctx, &ad_v1.ListAdsRequest{})

			require.Error(t, err)
			assert.Empty(t, resp.GetAds())

			st, ok := status.FromError(err)
			require.True(t, ok, "expected a gRPC status error")
			assert.Equal(t, tt.expectedCode, st.Code())
			assert.Contains(t, st.Message(), tt.expectedError)
		})
	}
}
