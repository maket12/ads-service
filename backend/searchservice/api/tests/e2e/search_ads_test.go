///go:build e2e

package e2e

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/maket12/ads-service/backend/authservice/pkg/utils"
	"github.com/maket12/ads-service/backend/searchservice/api/proto/generated/search_v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//func TestSearchAds_Success(t *testing.T) {
//	app := setupE2E(t)
//}

func TestSearchAds_BadCases(t *testing.T) {
	app := setupE2E(t)

	type testCase struct {
		name          string
		text          string
		category      *string
		priceFrom     *int64
		priceTo       *int64
		limit         int32
		offset        int32
		sortBy        string
		expectedCode  codes.Code
		expectedError string
	}

	var tests = []testCase{
		{
			name:          "Invalid Argument - Unknown Category",
			text:          gofakeit.HipsterSentence(),
			category:      utils.VPtr("unsupported"),
			expectedCode:  codes.InvalidArgument,
			expectedError: "invalid input",
		},
		{
			name:          "Invalid Argument - Negative PriceFrom",
			text:          gofakeit.HipsterSentence(),
			priceFrom:     utils.VPtr(int64(-100)),
			expectedCode:  codes.InvalidArgument,
			expectedError: "invalid input",
		},
		{
			name:          "Invalid Argument - Negative PriceTo",
			text:          gofakeit.HipsterSentence(),
			priceFrom:     utils.VPtr(int64(1000)),
			priceTo:       utils.VPtr(int64(-100)),
			expectedCode:  codes.InvalidArgument,
			expectedError: "invalid input",
		},
		{
			name:          "Invalid Argument - Negative Price Range",
			text:          gofakeit.HipsterSentence(),
			priceFrom:     utils.VPtr(int64(1000)),
			priceTo:       utils.VPtr(int64(100)),
			expectedCode:  codes.InvalidArgument,
			expectedError: "invalid input",
		},
		{
			name:          "Invalid Argument - Unknown Sort Option",
			text:          gofakeit.ProductName(),
			sortBy:        gofakeit.BeerAlcohol(),
			expectedCode:  codes.InvalidArgument,
			expectedError: "invalid input",
		},
		{
			name:          "Unauthenticated - missing account id",
			text:          gofakeit.ProductName(),
			category:      utils.VPtr("travel"),
			expectedCode:  codes.Unauthenticated,
			expectedError: "you must be authenticated to make this request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := app.userCtx()
			if tt.expectedCode == codes.Unauthenticated {
				ctx = context.Background()
			}

			resp, err := app.client.SearchAds(ctx, &search_v1.SearchAdsRequest{
				Text:      tt.text,
				Category:  tt.category,
				PriceFrom: tt.priceFrom,
				PriceTo:   tt.priceTo,
				Limit:     tt.limit,
				Offset:    tt.offset,
				SortBy:    tt.sortBy,
			})

			require.Error(t, err)
			require.Nil(t, resp.GetItems())
			require.Zero(t, resp.GetTotal())

			st, ok := status.FromError(err)
			require.True(t, ok, "expected a gRPC status error")
			assert.Equal(t, tt.expectedCode, st.Code())
			assert.Contains(t, st.Message(), tt.expectedError)
		})
	}
}
