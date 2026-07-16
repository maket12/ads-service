///go:build e2e

package e2e

//func TestVerifyEmail_Success(t *testing.T) {
//	app := setupE2E(t)
//	ctx := context.Background()
//	accountID, _, _ := app.createAccount(t, nil, nil, nil, nil, true)
//
//	t.Run("Successfully sent", func(t *testing.T) {
//		resp, err := app.client.SendVerification(ctx,
//			&auth_v1.SendVerificationRequest{
//				AccountId: accountID,
//			},
//		)
//		require.NoError(t, err)
//		require.True(t, resp.GetSent())
//	})
//}
//
//func TestVerifyEmail_BadCases(t *testing.T) {
//	app := setupE2E(t)
//	ctx := context.Background()
//
//	type testCase struct {
//		name          string
//		accountID     string
//		expectedCode  codes.Code
//		expectedError string
//	}
//
//	var tests = []testCase{
//		{
//			name:          "Not Found - Account Doesn't Exist",
//			accountID:     gofakeit.UUID(),
//			expectedCode:  codes.NotFound,
//			expectedError: "account not found",
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			resp, err := app.client.SendVerification(ctx,
//				&auth_v1.SendVerificationRequest{AccountId: tt.accountID},
//			)
//			require.Error(t, err)
//			assert.False(t, resp.GetSent())
//
//			st, ok := status.FromError(err)
//			require.True(t, ok, "expected a gRPC status error")
//			assert.Equal(t, tt.expectedCode, st.Code())
//			assert.Contains(t, st.Message(), tt.expectedError)
//		})
//	}
//}
