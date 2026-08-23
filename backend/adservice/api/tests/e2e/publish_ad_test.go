//go:build e2e

package e2e

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/maket12/ads-service/backend/adservice/pkg/generated/ad_v1"
	"github.com/maket12/ads-service/backend/adservice/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestPublishAd_Success(t *testing.T) {
	app := setupE2E(t)

	adID, sellerID := app.createAd(t, nil, nil)
	ctx := utils.PackAccountIDForGRPC(context.Background(), sellerID)

	resp, err := app.client.PublishAd(ctx, &ad_v1.PublishAdRequest{AdId: adID})

	require.NoError(t, err)
	require.True(t, resp.GetSuccess())
}

func TestPublishAd_BadCases(t *testing.T) {
	app := setupE2E(t)

	adID, sellerID := app.createAd(t, nil, nil)
	publishedAdID, _ := app.createAd(t, &sellerID, nil)
	app.publishAd(t, publishedAdID, sellerID)

	type testCase struct {
		name          string
		sellerID      string
		adID          string
		expectedCode  codes.Code
		expectedError string
	}

	var tests = []testCase{
		{
			name:          "Not Found - Ad Doesn't Exist",
			sellerID:      sellerID,
			adID:          gofakeit.UUID(),
			expectedCode:  codes.NotFound,
			expectedError: "ad not found",
		},
		{
			name:          "Failed Precondition - Ad Has Been Published",
			sellerID:      sellerID,
			adID:          publishedAdID,
			expectedCode:  codes.FailedPrecondition,
			expectedError: "ad has been already published or not available",
		},
		{
			name:          "Unauthenticated - missing seller id",
			sellerID:      "",
			adID:          adID,
			expectedCode:  codes.Unauthenticated,
			expectedError: "you must be authenticated to make this request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := utils.PackAccountIDForGRPC(context.Background(), tt.sellerID)

			resp, err := app.client.PublishAd(ctx, &ad_v1.PublishAdRequest{
				AdId: tt.adID,
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
