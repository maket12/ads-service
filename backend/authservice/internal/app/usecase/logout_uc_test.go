package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/maket12/ads-service/authservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/authservice/internal/app/errs"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	"github.com/maket12/ads-service/authservice/internal/domain/port"
	"github.com/maket12/ads-service/authservice/internal/domain/port/mocks"
	pkgerrs "github.com/maket12/ads-service/authservice/pkg/errs"
	"github.com/maket12/ads-service/authservice/pkg/utils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLogoutUC_Execute(t *testing.T) {
	type adapter struct {
		refreshSession *mocks.MockRefreshSessionRepository
		tokenGenerator *mocks.MockTokenGenerator
	}

	type testCase struct {
		name          string
		input         dto.LogoutInput
		mockBehaviour func(a adapter, accountID, sessionID uuid.UUID, rawToken string)
		expectErr     error
	}

	rawToken := "valid-refresh-token"
	ttl := 24 * time.Hour

	var tests = []testCase{
		{
			name: "Success",
			input: dto.LogoutInput{
				RefreshToken: rawToken,
			},
			mockBehaviour: func(a adapter, accountID, sessionID uuid.UUID, rawToken string) {
				session, _ := model.NewRefreshSession(sessionID, accountID, utils.HashToken(rawToken), nil, nil, nil, ttl)

				a.tokenGenerator.EXPECT().
					ValidateRefreshToken(mock.Anything, rawToken).
					Return(accountID, sessionID, nil)

				a.refreshSession.EXPECT().
					GetByID(mock.Anything, sessionID).
					Return(session, nil)

				a.refreshSession.EXPECT().
					Update(mock.Anything, session).
					Return(nil)
			},
			expectErr: nil,
		},
		{
			name: "Success - token expired but allowed to logout",
			input: dto.LogoutInput{
				RefreshToken: rawToken,
			},
			mockBehaviour: func(a adapter, accountID, sessionID uuid.UUID, rawToken string) {
				session, _ := model.NewRefreshSession(sessionID, accountID, utils.HashToken(rawToken), nil, nil, nil, ttl)

				a.tokenGenerator.EXPECT().
					ValidateRefreshToken(mock.Anything, rawToken).
					Return(accountID, sessionID, port.ErrTokenExpired)

				a.refreshSession.EXPECT().
					GetByID(mock.Anything, sessionID).
					Return(session, nil)

				a.refreshSession.EXPECT().
					Update(mock.Anything, session).
					Return(nil)
			},
			expectErr: nil,
		},
		{
			name: "Failure - token validation critical error",
			input: dto.LogoutInput{
				RefreshToken: rawToken,
			},
			mockBehaviour: func(a adapter, accountID, sessionID uuid.UUID, rawToken string) {
				a.tokenGenerator.EXPECT().
					ValidateRefreshToken(mock.Anything, rawToken).
					Return(uuid.Nil, uuid.Nil, errors.New("invalid signature"))
			},
			expectErr: ucerrs.ErrInvalidRefreshToken,
		},
		{
			name: "Failure - session not found",
			input: dto.LogoutInput{
				RefreshToken: rawToken,
			},
			mockBehaviour: func(a adapter, accountID, sessionID uuid.UUID, rawToken string) {
				a.tokenGenerator.EXPECT().
					ValidateRefreshToken(mock.Anything, rawToken).
					Return(accountID, sessionID, nil)

				a.refreshSession.EXPECT().
					GetByID(mock.Anything, sessionID).
					Return(nil, pkgerrs.ErrObjectNotFound)
			},
			expectErr: ucerrs.ErrInvalidRefreshToken,
		},
		{
			name: "Failure - session db error",
			input: dto.LogoutInput{
				RefreshToken: rawToken,
			},
			mockBehaviour: func(a adapter, accountID, sessionID uuid.UUID, rawToken string) {
				a.tokenGenerator.EXPECT().
					ValidateRefreshToken(mock.Anything, rawToken).
					Return(accountID, sessionID, nil)

				a.refreshSession.EXPECT().
					GetByID(mock.Anything, sessionID).
					Return(nil, errors.New("internal db error"))
			},
			expectErr: ucerrs.ErrGetRefreshSessionByIDDB,
		},
		{
			name: "Failure - session already inactive",
			input: dto.LogoutInput{
				RefreshToken: rawToken,
			},
			mockBehaviour: func(a adapter, accountID, sessionID uuid.UUID, rawToken string) {
				session, _ := model.NewRefreshSession(sessionID, accountID, utils.HashToken(rawToken), nil, nil, nil, ttl)
				_ = session.RevokeByLogout()

				a.tokenGenerator.EXPECT().
					ValidateRefreshToken(mock.Anything, rawToken).
					Return(accountID, sessionID, nil)

				a.refreshSession.EXPECT().
					GetByID(mock.Anything, sessionID).
					Return(session, nil)
			},
			expectErr: ucerrs.ErrInvalidRefreshToken,
		},
		{
			name: "Failure - token hash mismatch",
			input: dto.LogoutInput{
				RefreshToken: rawToken,
			},
			mockBehaviour: func(a adapter, accountID, sessionID uuid.UUID, rawToken string) {
				session, _ := model.NewRefreshSession(sessionID, accountID, utils.HashToken("different-token"), nil, nil, nil, ttl)

				a.tokenGenerator.EXPECT().
					ValidateRefreshToken(mock.Anything, rawToken).
					Return(accountID, sessionID, nil)

				a.refreshSession.EXPECT().
					GetByID(mock.Anything, sessionID).
					Return(session, nil)
			},
			expectErr: ucerrs.ErrInvalidRefreshToken,
		},
		{
			name: "Failure - update db error",
			input: dto.LogoutInput{
				RefreshToken: rawToken,
			},
			mockBehaviour: func(a adapter, accountID, sessionID uuid.UUID, rawToken string) {
				session, _ := model.NewRefreshSession(sessionID, accountID, utils.HashToken(rawToken), nil, nil, nil, ttl)

				a.tokenGenerator.EXPECT().
					ValidateRefreshToken(mock.Anything, rawToken).
					Return(accountID, sessionID, nil)

				a.refreshSession.EXPECT().
					GetByID(mock.Anything, sessionID).
					Return(session, nil)

				a.refreshSession.EXPECT().
					Update(mock.Anything, session).
					Return(errors.New("db save failure"))
			},
			expectErr: ucerrs.ErrRevokeRefreshSessionDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountID := uuid.New()
			sessionID := uuid.New()

			refreshSessionRepo := mocks.NewMockRefreshSessionRepository(t)
			tokenGenerator := mocks.NewMockTokenGenerator(t)

			tt.mockBehaviour(adapter{
				refreshSession: refreshSessionRepo,
				tokenGenerator: tokenGenerator,
			}, accountID, sessionID, rawToken)

			uc := NewLogoutUC(refreshSessionRepo, tokenGenerator)

			out, err := uc.Execute(context.Background(), tt.input)

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
				assert.False(t, out.Logout)
			} else {
				assert.NoError(t, err)
				assert.True(t, out.Logout)
			}
		})
	}
}
