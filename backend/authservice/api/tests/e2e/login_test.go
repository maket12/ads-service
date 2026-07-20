//go:build e2e

package e2e

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/maket12/ads-service/authservice/pkg/generated/auth_v1"
	"github.com/maket12/ads-service/authservice/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLogin_Success(t *testing.T) {
	app := setupE2E(t)
	ctx := context.Background()

	// Create a new account in advance
	email := gofakeit.Email()
	password := gofakeit.Password(true, true, true, true, true, 10)
	_, _, _ = app.createAccount(t,
		&email, &password,
		nil, nil, false,
	)

	resp, err := app.client.Login(ctx, &auth_v1.LoginRequest{
		Email:     email,
		Password:  password,
		Ip:        utils.VPtr(gofakeit.IPv4Address()),
		UserAgent: utils.VPtr(gofakeit.UserAgent()),
	})

	require.NoError(t, err)
	require.NotEmpty(t, resp.GetAccessToken())
	require.NotEmpty(t, resp.GetRefreshToken())
}

func TestLogin_BadCases(t *testing.T) {
	app := setupE2E(t)
	ctx := context.Background()

	/*
		Prepare test data:
		1) Create an account in advance for checking invalid credentials case
		2) Create another account and block it
		3) Create the third account and delete it
	*/
	existingEmail, blockedEmail, deletedEmail := gofakeit.Email(), gofakeit.Email(), gofakeit.Email()
	password := gofakeit.Password(true, true, true, true, true, 10)

	_, _, _ = app.createAccount(t, &existingEmail, &password, nil, nil, false)
	blockedAccID, _, _ := app.createAccount(t, &blockedEmail, &password, nil, nil, false)
	deletedAccID, _, _ := app.createAccount(t, &deletedEmail, &password, nil, nil, false)

	app.blockAccount(t, blockedAccID)
	app.deleteAccount(t, deletedAccID)

	type testCase struct {
		name          string
		email         string
		password      string
		expectedCode  codes.Code
		expectedError string
	}

	var tests = []testCase{
		{
			name:          "Not Found - Account Not Found",
			email:         gofakeit.Email(),
			password:      gofakeit.Password(true, true, true, true, true, 10),
			expectedCode:  codes.NotFound,
			expectedError: "invalid email or password",
		},
		{
			name:          "Not Found - Wrong Password",
			email:         existingEmail,
			password:      gofakeit.Password(true, true, true, true, true, 10),
			expectedCode:  codes.NotFound,
			expectedError: "invalid email or password",
		},
		{
			name:          "Failed Precondition - Account Is Blocked",
			email:         blockedEmail,
			password:      password,
			expectedCode:  codes.FailedPrecondition,
			expectedError: "account is either blocked or not exists",
		},
		{
			name:          "Failed Precondition - Account Is Deleted",
			email:         deletedEmail,
			password:      password,
			expectedCode:  codes.FailedPrecondition,
			expectedError: "account is either blocked or not exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := app.client.Login(ctx, &auth_v1.LoginRequest{
				Email:     tt.email,
				Password:  tt.password,
				Ip:        utils.VPtr(gofakeit.IPv4Address()),
				UserAgent: utils.VPtr(gofakeit.UserAgent()),
			})

			require.Error(t, err)
			assert.Empty(t, resp.GetAccessToken())
			assert.Empty(t, resp.GetRefreshToken())

			st, ok := status.FromError(err)
			require.True(t, ok, "expected a gRPC status error")
			assert.Equal(t, tt.expectedCode, st.Code())
			assert.Contains(t, st.Message(), tt.expectedError)
		})
	}
}
