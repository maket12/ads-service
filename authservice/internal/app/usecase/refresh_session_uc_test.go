package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/maket12/ads-service/authservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/authservice/internal/app/errs"
	"github.com/maket12/ads-service/authservice/internal/app/usecase"
	"github.com/maket12/ads-service/authservice/internal/app/utils"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	"github.com/maket12/ads-service/authservice/internal/domain/port/mocks"
	pkgerrs "github.com/maket12/ads-service/pkg/errs"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRefreshSessionUC_Execute(t *testing.T) {
	type adapter struct {
		accountRole    *mocks.AccountRoleRepository
		refreshSession *mocks.RefreshSessionRepository
		tokenGenerator *mocks.TokenGenerator
	}

	type testCase struct {
		name    string
		input   dto.RefreshSessionInput
		prepare func(a adapter)
		wantErr error
	}

	accountID := uuid.New()
	oldSessionID := uuid.New()
	oldToken := "old-refresh-token"
	hashedOldToken := utils.HashToken(oldToken)
	ip := "127.0.0.1"
	anotherIP := "8.8.8.8"
	ua := "Mozilla/5.0"
	anotherUA := "HackerBrowser/1.0"
	ttl := time.Hour * 24

	activeOldSession, _ := model.NewRefreshSession(
		oldSessionID, accountID, hashedOldToken, nil, &ip, &ua, ttl,
	)

	role, _ := model.NewAccountRole(accountID)

	var tests = []testCase{
		{
			name: "Success - Token Rotation",
			input: dto.RefreshSessionInput{
				OldRefreshToken: oldToken,
				IP:              &ip,
				UserAgent:       &ua,
			},
			prepare: func(a adapter) {
				a.tokenGenerator.On("ValidateRefreshToken", mock.Anything, oldToken).
					Return(accountID, oldSessionID, nil)

				a.refreshSession.On("GetByID", mock.Anything, oldSessionID).
					Return(activeOldSession, nil)

				a.refreshSession.On("Revoke", mock.Anything, mock.MatchedBy(func(s *model.RefreshSession) bool {
					return !s.IsActive() && s.ID() == oldSessionID
				})).Return(nil)

				a.accountRole.On("Get", mock.Anything, accountID).Return(role, nil)

				a.tokenGenerator.On("GenerateAccessToken", mock.Anything, accountID, "user").
					Return("new-access-token", nil)
				a.tokenGenerator.On("GenerateRefreshToken", mock.Anything, accountID, mock.Anything).
					Return("new-refresh-token", nil)

				a.refreshSession.On("Create", mock.Anything, mock.MatchedBy(func(s *model.RefreshSession) bool {
					return s.RotatedFrom() != nil && *s.RotatedFrom() == oldSessionID
				})).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "Fail - IP Mismatch",
			input: dto.RefreshSessionInput{
				OldRefreshToken: oldToken,
				IP:              &anotherIP,
				UserAgent:       &ua,
			},
			prepare: func(a adapter) {
				a.tokenGenerator.On("ValidateRefreshToken", mock.Anything, oldToken).
					Return(accountID, oldSessionID, nil)
				a.refreshSession.On("GetByID", mock.Anything, oldSessionID).
					Return(activeOldSession, nil)
			},
			wantErr: ucerrs.ErrInvalidRefreshToken,
		},
		{
			name: "Fail - UserAgent Mismatch",
			input: dto.RefreshSessionInput{
				OldRefreshToken: oldToken,
				IP:              &ip,
				UserAgent:       &anotherUA,
			},
			prepare: func(a adapter) {
				a.tokenGenerator.On("ValidateRefreshToken", mock.Anything, oldToken).
					Return(accountID, oldSessionID, nil)
				a.refreshSession.On("GetByID", mock.Anything, oldSessionID).
					Return(activeOldSession, nil)
			},
			wantErr: ucerrs.ErrInvalidRefreshToken,
		},
		{
			name: "Fail - Old Token Hash Mismatch",
			input: dto.RefreshSessionInput{
				OldRefreshToken: "wrong-token-for-this-session",
				IP:              &ip,
				UserAgent:       &ua,
			},
			prepare: func(a adapter) {
				a.tokenGenerator.On("ValidateRefreshToken", mock.Anything, "wrong-token-for-this-session").
					Return(accountID, oldSessionID, nil)
				a.refreshSession.On("GetByID", mock.Anything, oldSessionID).
					Return(activeOldSession, nil)
			},
			wantErr: ucerrs.ErrInvalidRefreshToken,
		},
		{
			name:  "Fail - Session Not Found",
			input: dto.RefreshSessionInput{OldRefreshToken: oldToken, IP: &ip, UserAgent: &ua},
			prepare: func(a adapter) {
				a.tokenGenerator.On("ValidateRefreshToken", mock.Anything, oldToken).
					Return(accountID, oldSessionID, nil)
				a.refreshSession.On("GetByID", mock.Anything, oldSessionID).
					Return(nil, pkgerrs.ErrObjectNotFound)
			},
			wantErr: ucerrs.ErrInvalidRefreshToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := adapter{
				accountRole:    mocks.NewAccountRoleRepository(t),
				refreshSession: mocks.NewRefreshSessionRepository(t),
				tokenGenerator: mocks.NewTokenGenerator(t),
			}

			tt.prepare(a)

			uc := usecase.NewRefreshSessionUC(a.accountRole, a.refreshSession, a.tokenGenerator, ttl)

			res, err := uc.Execute(context.Background(), tt.input)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, res.AccessToken)
				assert.NotEmpty(t, res.RefreshToken)
			}
		})
	}
}
