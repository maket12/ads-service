package e2e

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/maket12/ads-service/userservice/pkg/generated/user_v1"
	"github.com/maket12/ads-service/userservice/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUpdateProfile_Success(t *testing.T) {
	app := setupE2E(t)
	accountID := app.createProfile(t, nil)
	ctx := utils.PackAccountIDForGRPC(context.Background(), accountID)

	var (
		fName  = gofakeit.FirstName()
		lName  = gofakeit.LastName()
		phone  = gofakeit.Phone()
		avatar = gofakeit.URL()
		bio    = gofakeit.Bio()
	)

	resp, err := app.client.UpdateProfile(ctx, &user_v1.UpdateProfileRequest{
		FirstName: &fName,
		LastName:  &lName,
		Phone:     &phone,
		AvatarUrl: &avatar,
		Bio:       &bio,
	})

	require.NoError(t, err)
	require.True(t, resp.GetSuccess())

	// Ensure the profile was updated
	profile := app.getProfile(t, accountID)

	assert.Equal(t, fName, profile.GetFirstName())
	assert.Equal(t, lName, profile.GetLastName())
	assert.Equal(t, phone, profile.GetPhone())
	assert.Equal(t, avatar, profile.GetAvatarUrl())
	assert.Equal(t, bio, profile.GetBio())
}

func TestUpdateProfile_BadCases(t *testing.T) {
	app := setupE2E(t)
	accountID := app.createProfile(t, nil)

	type testCase struct {
		name          string
		accountID     string
		firstName     *string
		lastName      *string
		phone         *string
		avatarURL     *string
		bio           *string
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
		{
			name:          "Invalid Argument - Phone",
			accountID:     accountID,
			phone:         utils.VPtr(gofakeit.Phone()[3:]),
			expectedCode:  codes.InvalidArgument,
			expectedError: "invalid input",
		},
		{
			name: "Invalid Argument - Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := utils.PackAccountIDForGRPC(context.Background(), tt.accountID)
			resp, err := app.client.UpdateProfile(ctx, &user_v1.UpdateProfileRequest{
				FirstName: tt.firstName,
				LastName:  tt.lastName,
				Phone:     tt.phone,
				AvatarUrl: tt.avatarURL,
				Bio:       tt.bio,
			})

			require.Error(t, err)
			require.False(t, resp.GetSuccess())

			st, ok := status.FromError(err)
			require.True(t, ok, "expected a gRPC status error")
			assert.Equal(t, tt.expectedCode, st.Code())
			assert.Contains(t, st.Message(), tt.expectedError)
		})
	}
}
