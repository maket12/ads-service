///go:build e2e

package e2e

import (
	"context"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/maket12/ads-service/adservice/pkg/generated/ad_v1"
	"github.com/maket12/ads-service/adservice/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUpdateAd_Success(t *testing.T) {
	app := setupE2E(t)

	adID, sellerID := app.createAd(t, nil, nil)
	ctx := utils.PackAccountIDForGRPC(context.Background(), sellerID)

	resp, err := app.client.UpdateAd(ctx, &ad_v1.UpdateAdRequest{
		AdId:  adID,
		Title: utils.VPtr(gofakeit.ProductName()),
	})

	require.NoError(t, err)
	require.True(t, resp.GetSuccess())
}

func TestUpdateAd_BadCases(t *testing.T) {
	app := setupE2E(t)

	adID, sellerID := app.createAd(t, nil, nil)
	deletedAdID, _ := app.createAd(t, &sellerID, nil)
	app.publishAd(t, deletedAdID, sellerID)
	app.deleteAd(t, deletedAdID, sellerID)

	type testCase struct {
		name          string
		sellerID      string
		adID          string
		title         *string
		description   *string
		price         *int64
		expectedCode  codes.Code
		expectedError string
	}

	var tests = []testCase{
		{
			name:          "Invalid Argument - Title Is Short",
			sellerID:      sellerID,
			adID:          adID,
			title:         utils.VPtr("tit"),
			expectedCode:  codes.InvalidArgument,
			expectedError: "invalid input",
		},
		{
			name:          "Invalid Argument - Description Is Too Long",
			sellerID:      sellerID,
			adID:          adID,
			description:   utils.VPtr(strings.Repeat(gofakeit.ProductDescription(), 100)),
			expectedCode:  codes.InvalidArgument,
			expectedError: "invalid input",
		},
		{
			name:          "Invalid Argument - Negative Price",
			sellerID:      sellerID,
			adID:          adID,
			price:         utils.VPtr(int64(-100)),
			expectedCode:  codes.InvalidArgument,
			expectedError: "invalid input",
		},
		{
			name:          "Not Found - Ad Doesn't Exist",
			sellerID:      sellerID,
			adID:          gofakeit.UUID(),
			expectedCode:  codes.NotFound,
			expectedError: "ad not found",
		},
		{
			name:          "Permission Denied - Account Isn't Owner",
			sellerID:      gofakeit.UUID(),
			adID:          adID,
			expectedCode:  codes.PermissionDenied,
			expectedError: "no permission to access this data",
		},
		{
			name:          "Failed Precondition - Ad Isn't Available",
			sellerID:      sellerID,
			adID:          deletedAdID,
			expectedCode:  codes.FailedPrecondition,
			expectedError: "ad cannot be updated",
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

			resp, err := app.client.UpdateAd(ctx, &ad_v1.UpdateAdRequest{
				AdId:        tt.adID,
				Title:       tt.title,
				Description: tt.description,
				Price:       tt.price,
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
