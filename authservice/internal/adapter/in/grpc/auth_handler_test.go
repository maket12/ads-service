package grpc_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/maket12/ads-service/authservice/internal/adapter/in/grpc"
	"github.com/maket12/ads-service/authservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/authservice/internal/app/errs"
	"github.com/maket12/ads-service/authservice/internal/app/usecase/mocks"
	"github.com/maket12/ads-service/pkg/generated/auth_v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
)

func TestAH_Register(t *testing.T) {
	testUID := uuid.New()

	type testCase struct {
		name      string
		request   *auth_v1.RegisterRequest
		setupMock func(m *mocks.RegisterUseCase)
		wantCode  codes.Code
		wantResp  *auth_v1.RegisterResponse
	}

	testCases := []testCase{
		{
			name: "Success registration",
			request: &auth_v1.RegisterRequest{
				Email:    "shishi12377@weixin.cn",
				Password: "liushi07.12.2006",
			},
			setupMock: func(m *mocks.RegisterUseCase) {
				m.On("Execute", mock.Anything, mock.Anything).
					Return(dto.RegisterOutput{AccountID: testUID}, nil)
			},
			wantCode: codes.OK,
			wantResp: &auth_v1.RegisterResponse{AccountId: testUID.String()},
		},
		{
			name: "Failure - invalid input",
			request: &auth_v1.RegisterRequest{
				Email:    "",
				Password: "",
			},
			setupMock: func(m *mocks.RegisterUseCase) {
				m.On("Execute", mock.Anything, mock.Anything).
					Return(dto.RegisterOutput{AccountID: uuid.Nil}, ucerrs.ErrInvalidInput)
			},
			wantCode: codes.InvalidArgument,
			wantResp: &auth_v1.RegisterResponse{AccountId: uuid.Nil.String()},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mockReg := mocks.NewRegisterUseCase(t)
			if tt.setupMock != nil {
				tt.setupMock(mockReg)
			}

			handler := grpc.NewAuthHandler(slog.Default(), mockReg, nil,
				nil, nil, nil, nil,
			)

			resp, err := handler.Register(context.Background(), tt.request)

			if tt.wantCode == codes.OK {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResp, resp)
			}
		})
	}
}
