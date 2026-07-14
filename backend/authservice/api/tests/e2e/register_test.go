///go:build e2e

package e2e

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/maket12/ads-service/authservice/pkg/generated/auth_v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRegister_Success(t *testing.T) {
	app := setupE2E(t)
	ctx := context.Background()

	resp, err := app.client.Register(ctx, &auth_v1.RegisterRequest{
		Email:    gofakeit.Email(),
		Password: gofakeit.Password(true, true, true, true, true, 10),
	})

	require.NoError(t, err)
	require.NotEmpty(t, resp.GetAccountId())
}

func TestRegister_BadCases(t *testing.T) {
	app := setupE2E(t)
	ctx := context.Background()

	// Create an account in advance for checking email duplicate case
	existingEmail := gofakeit.Email()
	_ = app.createAccount(t, &existingEmail, nil)

	type testCase struct {
		name          string
		email         string
		password      string
		expectedCode  codes.Code
		expectedError string
	}

	var tests = []testCase{
		{
			name:          "Invalid Argument - Email Not Specified",
			email:         "",
			password:      gofakeit.Password(true, true, true, true, true, 10),
			expectedCode:  codes.InvalidArgument,
			expectedError: "invalid input",
		},
		{
			name:          "Invalid Argument - Password Not Specified",
			email:         gofakeit.Email(),
			password:      "",
			expectedCode:  codes.InvalidArgument,
			expectedError: "invalid input",
		},
		{
			name:          "Invalid Argument - Invalid Email",
			email:         "not-a-email",
			password:      gofakeit.Password(true, true, true, true, true, 10),
			expectedCode:  codes.InvalidArgument,
			expectedError: "invalid input",
		},
		{
			name:          "Already Exists - Duplicate Email",
			email:         existingEmail,
			password:      gofakeit.Password(true, true, true, true, true, 10),
			expectedCode:  codes.AlreadyExists,
			expectedError: "account with given email already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := app.client.Register(ctx, &auth_v1.RegisterRequest{
				Email:    tt.email,
				Password: tt.password,
			})

			require.Error(t, err)
			assert.Empty(t, resp.GetAccountId())

			st, ok := status.FromError(err)
			require.True(t, ok, "expected a gRPC status error")
			assert.Equal(t, tt.expectedCode, st.Code())
			assert.Contains(t, st.Message(), tt.expectedError)
		})
	}
}
