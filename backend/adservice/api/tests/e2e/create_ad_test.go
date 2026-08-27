//go:build e2e

package e2e

import (
	"context"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/maket12/ads-service/backend/adservice/pkg/generated/ad_v1"
	"github.com/maket12/ads-service/backend/adservice/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateAd_Success(t *testing.T) {
	app := setupE2E(t)
	sellerID := gofakeit.UUID()
	ctx := utils.PackAccountIDForGRPC(context.Background(), sellerID)

	resp, err := app.client.CreateAd(ctx, &ad_v1.CreateAdRequest{
		Title:       gofakeit.BookTitle(),
		Description: utils.VPtr(gofakeit.ProductDescription()),
		Price:       gofakeit.Int64(),
		Category:    "food",
		Images:      []string{gofakeit.URL(), gofakeit.URL()},
	})

	require.NoError(t, err)
	require.NotEmpty(t, resp.GetId())
}

func TestCreateAd_BadCases(t *testing.T) {
	app := setupE2E(t)

	type testCase struct {
		name          string
		sellerID      string
		title         string
		description   *string
		price         int64
		category      string
		images        []string
		expectedCode  codes.Code
		expectedError string
	}

	var tests = []testCase{
		{
			name:          "Invalid Argument - Seller ID",
			sellerID:      uuid.Nil.String(),
			title:         gofakeit.ProductName(),
			description:   utils.VPtr(gofakeit.ProductDescription()),
			price:         10000,
			category:      "food",
			images:        nil,
			expectedCode:  codes.InvalidArgument,
			expectedError: "invalid input",
		},
		{
			name:          "Invalid Argument - Title Isn't Specified",
			sellerID:      gofakeit.UUID(),
			title:         "",
			description:   utils.VPtr(gofakeit.ProductDescription()),
			price:         20000,
			category:      "food",
			images:        nil,
			expectedCode:  codes.InvalidArgument,
			expectedError: "invalid input",
		},
		{
			name:          "Invalid Argument - Description's Out Of Boundaries",
			sellerID:      gofakeit.UUID(),
			title:         gofakeit.ProductName(),
			description:   utils.VPtr(strings.Repeat(gofakeit.ProductDescription(), 100)),
			price:         300,
			category:      "food",
			images:        nil,
			expectedCode:  codes.InvalidArgument,
			expectedError: "invalid input",
		},
		{
			name:          "Invalid Argument - Negative Price",
			sellerID:      gofakeit.UUID(),
			title:         gofakeit.ProductName(),
			description:   utils.VPtr(gofakeit.ProductDescription()),
			price:         -100,
			category:      "food",
			images:        nil,
			expectedCode:  codes.InvalidArgument,
			expectedError: "invalid input",
		},
		{
			name:          "Invalid Argument - Unknown Category",
			sellerID:      gofakeit.UUID(),
			title:         gofakeit.ProductName(),
			description:   utils.VPtr(gofakeit.ProductDescription()),
			price:         100,
			category:      "unsupported",
			images:        nil,
			expectedCode:  codes.InvalidArgument,
			expectedError: "invalid input",
		},
		{
			name:          "Unauthenticated - missing seller id",
			sellerID:      "",
			title:         gofakeit.ProductName(),
			description:   utils.VPtr(gofakeit.ProductDescription()),
			price:         10000,
			category:      "food",
			images:        nil,
			expectedCode:  codes.Unauthenticated,
			expectedError: "you must be authenticated to make this request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := utils.PackAccountIDForGRPC(context.Background(), tt.sellerID)
			resp, err := app.client.CreateAd(ctx, &ad_v1.CreateAdRequest{
				Title:       tt.title,
				Description: tt.description,
				Price:       tt.price,
				Images:      tt.images,
			})

			require.Error(t, err)
			assert.Empty(t, resp.GetId())

			st, ok := status.FromError(err)
			require.True(t, ok, "expected a gRPC status error")
			assert.Equal(t, tt.expectedCode, st.Code())
			assert.Contains(t, st.Message(), tt.expectedError)
		})
	}
}
