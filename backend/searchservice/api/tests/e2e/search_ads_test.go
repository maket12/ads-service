//go:build e2e

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

func TestSearchAds_Success(t *testing.T) {
	app := setupE2E(t)

	id1 := app.createAdIndex(t, nil, utils.VPtr("Vintage Camera"), nil, utils.VPtr(int64(500)), utils.VPtr("electronics"), nil)
	id2 := app.createAdIndex(t, nil, utils.VPtr("Digital Camera DSLR"), nil, utils.VPtr(int64(1500)), utils.VPtr("electronics"), nil)
	id3 := app.createAdIndex(t, nil, utils.VPtr("Vegan Burger"), nil, utils.VPtr(int64(50)), utils.VPtr("food"), nil)
	id4 := app.createAdIndex(t, nil, utils.VPtr("Gaming Console 5"), nil, utils.VPtr(int64(3000)), utils.VPtr("video_games"), nil)

	type testCase struct {
		name          string
		req           *search_v1.SearchAdsRequest
		expectedTotal int64
		expectedIDs   []string
	}

	tests := []testCase{
		{
			name: "Full-text search by title - match 'Camera'",
			req: &search_v1.SearchAdsRequest{
				Text: "Camera",
			},
			expectedTotal: 2,
			expectedIDs:   []string{id1, id2},
		},
		{
			name: "Filter by exact category - 'food'",
			req: &search_v1.SearchAdsRequest{
				Category: utils.VPtr("food"),
			},
			expectedTotal: 1,
			expectedIDs:   []string{id3},
		},
		{
			name: "Filter by price range - 1000 to 4000",
			req: &search_v1.SearchAdsRequest{
				PriceFrom: utils.VPtr(int64(1000)),
				PriceTo:   utils.VPtr(int64(4000)),
			},
			expectedTotal: 2,
			expectedIDs:   []string{id2, id4},
		},
		{
			name: "Pagination - limit 1 offset 0 in electronics",
			req: &search_v1.SearchAdsRequest{
				Category: utils.VPtr("electronics"),
				Limit:    1,
				Offset:   0,
			},
			expectedTotal: 2,
		},
		{
			name: "Sort by price descending",
			req: &search_v1.SearchAdsRequest{
				SortBy: "price_desc",
			},
			expectedTotal: 4,
			expectedIDs:   []string{id4, id2, id1, id3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := app.client.SearchAds(app.userCtx(), tt.req)

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tt.expectedTotal, resp.GetTotal())

			if tt.req.Limit > 0 {
				assert.LessOrEqual(t, int32(len(resp.GetItems())), tt.req.Limit)
			}

			if len(tt.expectedIDs) > 0 && tt.req.Limit == 0 {
				require.Equal(t, len(tt.expectedIDs), len(resp.GetItems()))

				actualIDs := make([]string, 0, len(resp.GetItems()))
				for _, item := range resp.GetItems() {
					actualIDs = append(actualIDs, item.GetId())
				}

				if tt.req.SortBy != "" {
					assert.Equal(t, tt.expectedIDs, actualIDs)
				} else {
					assert.ElementsMatch(t, tt.expectedIDs, actualIDs)
				}
			}
		})
	}
}

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
