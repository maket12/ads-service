///go:build e2e

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
	_ = app.createAccount(t, &email, &password)

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

	// Create an account in advance for checking invalid credentials case
	existingEmail := gofakeit.Email()
	existingPassword := gofakeit.Password(true, true, true, true, true, 10)
	_ = app.createAccount(t, &existingEmail, &existingPassword)

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
