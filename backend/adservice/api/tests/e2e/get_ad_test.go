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

func TestGetAd_Success(t *testing.T) {
	app := setupE2E(t)

	payload := &ad_v1.CreateAdRequest{
		Title:       gofakeit.ProductName(),
		Description: utils.VPtr(gofakeit.ProductDescription()),
		Price:       int64(gofakeit.Price(10000, 50000)),
		Images:      []string{gofakeit.URL(), gofakeit.URL()},
	}
	adID, sellerID := app.createAd(t, nil, payload)

	ctx := utils.PackAccountIDForGRPC(context.Background(), sellerID)

	resp, err := app.client.GetAd(ctx, &ad_v1.GetAdRequest{Id: adID})
	require.NoError(t, err)

	require.NotEmpty(t, resp.GetAd().GetId())
	require.Equal(t, sellerID, resp.GetAd().GetSellerId())
	require.Equal(t, payload.GetTitle(), resp.GetAd().GetTitle())
	require.Equal(t, payload.GetDescription(), resp.GetAd().GetDescription())
	require.Equal(t, payload.GetPrice(), resp.GetAd().GetPrice())
	require.ElementsMatch(t, payload.GetImages(), resp.GetAd().GetImages())
}

func TestGetAd_BadCases(t *testing.T) {
	app := setupE2E(t)
	adID, sellerID := app.createAd(t, nil, nil)

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
			name:          "Permission Denied - Account Hasn't Access", // ad hasn't published
			sellerID:      gofakeit.UUID(),
			adID:          adID,
			expectedCode:  codes.PermissionDenied,
			expectedError: "no permission to access this data",
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

			resp, err := app.client.GetAd(ctx, &ad_v1.GetAdRequest{
				Id: tt.adID,
			})

			require.Error(t, err)
			assert.Empty(t, resp.GetAd().GetId())
			assert.Empty(t, resp.GetAd().GetSellerId())
			assert.Empty(t, resp.GetAd().GetTitle())
			assert.Empty(t, resp.GetAd().GetDescription())
			assert.Empty(t, resp.GetAd().GetPrice())
			assert.Empty(t, resp.GetAd().GetStatus())
			assert.Empty(t, resp.GetAd().GetImages())
			assert.Empty(t, resp.GetAd().GetCreatedAt())
			assert.Empty(t, resp.GetAd().GetUpdatedAt())

			st, ok := status.FromError(err)
			require.True(t, ok, "expected a gRPC status error")
			assert.Equal(t, tt.expectedCode, st.Code())
			assert.Contains(t, st.Message(), tt.expectedError)
		})
	}
}
