package usecase_test

import (
	"context"
	"github.com/maket12/ads-service/authservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/authservice/internal/app/errs"
	"github.com/maket12/ads-service/authservice/internal/app/usecase"
	"github.com/maket12/ads-service/authservice/internal/app/utils"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	"github.com/maket12/ads-service/authservice/internal/domain/port/mocks"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLogoutUC_Execute(t *testing.T) {
	type adapter struct {
		refreshSession *mocks.RefreshSessionRepository
		tokenGenerator *mocks.TokenGenerator
	}

	accountID := uuid.New()
	sessionID := uuid.New()
	token := "valid-refresh-token"
	hashedToken := utils.HashToken(token)

	activeSession, _ := model.NewRefreshSession(
		sessionID, accountID, hashedToken, nil, nil, nil, time.Hour,
	)

	type testCase struct {
		name    string
		input   dto.LogoutInput
		prepare func(a adapter)
		wantErr error
	}

	var tests = []testCase{
		{
			name: "Success",
			input: dto.LogoutInput{
				RefreshToken: token,
			},
			prepare: func(a adapter) {
				a.tokenGenerator.On("ValidateRefreshToken", mock.Anything, token).
					Return(accountID, sessionID, nil)

				a.refreshSession.On("GetByID", mock.Anything, sessionID).
					Return(activeSession, nil)

				a.refreshSession.On("Revoke", mock.Anything, mock.MatchedBy(func(s *model.RefreshSession) bool {
					return !s.IsActive()
				})).Return(nil)
			},
			wantErr: nil,
		},
		{
			name:  "Fail - Invalid Token Format",
			input: dto.LogoutInput{RefreshToken: "invalid"},
			prepare: func(a adapter) {
				a.tokenGenerator.On("ValidateRefreshToken", mock.Anything, "invalid").
					Return(uuid.Nil, uuid.Nil, ucerrs.ErrInvalidRefreshToken)
			},
			wantErr: ucerrs.ErrInvalidRefreshToken,
		},
		{
			name:  "Fail - Session Already Inactive",
			input: dto.LogoutInput{RefreshToken: token},
			prepare: func(a adapter) {
				reason := "prev logout"
				inactiveSession, _ := model.NewRefreshSession(sessionID, accountID, hashedToken, nil, nil, nil, time.Hour)
				_ = inactiveSession.Revoke(&reason)

				a.tokenGenerator.On("ValidateRefreshToken", mock.Anything, token).
					Return(accountID, sessionID, nil)
				a.refreshSession.On("GetByID", mock.Anything, sessionID).
					Return(inactiveSession, nil)
			},
			wantErr: ucerrs.ErrInvalidRefreshToken,
		},
		{
			name:  "Fail - Token Hash Mismatch",
			input: dto.LogoutInput{RefreshToken: "another-token"},
			prepare: func(a adapter) {
				a.tokenGenerator.On("ValidateRefreshToken", mock.Anything, "another-token").
					Return(accountID, sessionID, nil)
				a.refreshSession.On("GetByID", mock.Anything, sessionID).
					Return(activeSession, nil)
			},
			wantErr: ucerrs.ErrInvalidRefreshToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := adapter{
				refreshSession: mocks.NewRefreshSessionRepository(t),
				tokenGenerator: mocks.NewTokenGenerator(t),
			}

			if tt.prepare != nil {
				tt.prepare(a)
			}

			uc := usecase.NewLogoutUC(a.refreshSession, a.tokenGenerator)

			resp, err := uc.Execute(context.Background(), tt.input)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.False(t, resp.Logout)
			} else {
				assert.NoError(t, err)
				assert.True(t, resp.Logout)
			}
		})
	}

}
