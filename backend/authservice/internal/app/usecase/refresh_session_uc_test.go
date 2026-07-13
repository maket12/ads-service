package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/maket12/ads-service/authservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/authservice/internal/app/errs"
	"github.com/maket12/ads-service/authservice/internal/app/usecase"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	"github.com/maket12/ads-service/authservice/internal/domain/port"
	"github.com/maket12/ads-service/authservice/internal/domain/port/mocks"
	pkgerrs "github.com/maket12/ads-service/authservice/pkg/errs"
	"github.com/maket12/ads-service/authservice/pkg/utils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const refreshSessionTTL = 30 * 24 * time.Hour

func newTestSession(t *testing.T, id, accountID uuid.UUID, tokenHash string, ip, ua *string) *model.RefreshSession {
	t.Helper()
	s, err := model.NewRefreshSession(id, accountID, tokenHash, nil, ip, ua, refreshSessionTTL)
	assert.NoError(t, err)
	return s
}

func TestRefreshSessionUC_Execute(t *testing.T) {
	type adapter struct {
		accountRole    *mocks.MockAccountRoleRepository
		refreshSession *mocks.MockRefreshSessionRepository
		tokenGenerator *mocks.MockTokenGenerator
	}

	type testCase struct {
		name          string
		input         dto.RefreshSessionInput
		mockBehaviour func(a adapter, accountID, sessionID uuid.UUID, rawToken string)
		expectErr     error
	}

	rawToken := "raw-refresh-token"
	ip := utils.VPtr("1.2.3.4")
	ua := utils.VPtr("test-agent")

	var tests = []testCase{
		{
			name: "Success",
			input: dto.RefreshSessionInput{
				RefreshToken: rawToken,
				IP:           ip,
				UserAgent:    ua,
			},
			mockBehaviour: func(a adapter, accountID, sessionID uuid.UUID, rawToken string) {
				session := newTestSession(t, sessionID, accountID, utils.HashToken(rawToken), ip, ua)

				a.tokenGenerator.EXPECT().
					ValidateRefreshToken(mock.Anything, rawToken).
					Return(accountID, sessionID, nil)

				a.refreshSession.EXPECT().
					GetByID(mock.Anything, sessionID).
					Return(session, nil)

				a.refreshSession.EXPECT().
					Update(mock.Anything, session).
					Return(nil)

				a.accountRole.EXPECT().
					Get(mock.Anything, accountID).
					Return(model.RestoreAccountRole(accountID, model.RoleUser), nil)

				a.tokenGenerator.EXPECT().
					GeneratePair(mock.Anything, accountID, model.RoleUser.String(), mock.AnythingOfType("uuid.UUID")).
					Return(&port.TokensPair{Access: "new-access", Refresh: "new-refresh"}, nil)

				a.refreshSession.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*model.RefreshSession")).
					Return(nil)
			},
			expectErr: nil,
		},
		{
			name: "Failure - invalid token signature",
			input: dto.RefreshSessionInput{
				RefreshToken: rawToken,
				IP:           ip,
				UserAgent:    ua,
			},
			mockBehaviour: func(a adapter, accountID, sessionID uuid.UUID, rawToken string) {
				a.tokenGenerator.EXPECT().
					ValidateRefreshToken(mock.Anything, rawToken).
					Return(uuid.Nil, uuid.Nil, errors.New("bad signature"))
			},
			expectErr: ucerrs.ErrInvalidRefreshToken,
		},
		{
			name: "Failure - session not found",
			input: dto.RefreshSessionInput{
				RefreshToken: rawToken,
				IP:           ip,
				UserAgent:    ua,
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
			name: "Failure - GetByID db error",
			input: dto.RefreshSessionInput{
				RefreshToken: rawToken,
				IP:           ip,
				UserAgent:    ua,
			},
			mockBehaviour: func(a adapter, accountID, sessionID uuid.UUID, rawToken string) {
				a.tokenGenerator.EXPECT().
					ValidateRefreshToken(mock.Anything, rawToken).
					Return(accountID, sessionID, nil)

				a.refreshSession.EXPECT().
					GetByID(mock.Anything, sessionID).
					Return(nil, errors.New("connection reset"))
			},
			expectErr: ucerrs.ErrGetRefreshSessionByIDDB,
		},
		{
			name: "Failure - rotated-token reuse triggers descendant revocation",
			input: dto.RefreshSessionInput{
				RefreshToken: rawToken,
				IP:           ip,
				UserAgent:    ua,
			},
			mockBehaviour: func(a adapter, accountID, sessionID uuid.UUID, rawToken string) {
				session := newTestSession(t, sessionID, accountID, utils.HashToken(rawToken), ip, ua)
				assert.NoError(t, session.RevokeByRotation()) // simulate: already rotated once

				a.tokenGenerator.EXPECT().
					ValidateRefreshToken(mock.Anything, rawToken).
					Return(accountID, sessionID, nil)

				a.refreshSession.EXPECT().
					GetByID(mock.Anything, sessionID).
					Return(session, nil)

				a.refreshSession.EXPECT().
					RevokeDescendants(mock.Anything, sessionID, mock.AnythingOfType("*string")).
					Return(nil)
			},
			expectErr: ucerrs.ErrInvalidRefreshToken,
		},
		{
			name: "Failure - IP/UserAgent mismatch revokes as suspicious",
			input: dto.RefreshSessionInput{
				RefreshToken: rawToken,
				IP:           utils.VPtr("9.9.9.9"),
				UserAgent:    ua,
			},
			mockBehaviour: func(a adapter, accountID, sessionID uuid.UUID, rawToken string) {
				session := newTestSession(t, sessionID, accountID, utils.HashToken(rawToken), ip, ua)

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
			expectErr: ucerrs.ErrInvalidRefreshToken,
		},
		{
			name: "Failure - token hash mismatch",
			input: dto.RefreshSessionInput{
				RefreshToken: rawToken,
				IP:           ip,
				UserAgent:    ua,
			},
			mockBehaviour: func(a adapter, accountID, sessionID uuid.UUID, rawToken string) {
				// session stores a hash for a *different* token value
				session := newTestSession(t, sessionID, accountID, utils.HashToken("some-other-token"), ip, ua)

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
			name: "Failure - revoke old session db error",
			input: dto.RefreshSessionInput{
				RefreshToken: rawToken,
				IP:           ip,
				UserAgent:    ua,
			},
			mockBehaviour: func(a adapter, accountID, sessionID uuid.UUID, rawToken string) {
				session := newTestSession(t, sessionID, accountID, utils.HashToken(rawToken), ip, ua)

				a.tokenGenerator.EXPECT().
					ValidateRefreshToken(mock.Anything, rawToken).
					Return(accountID, sessionID, nil)

				a.refreshSession.EXPECT().
					GetByID(mock.Anything, sessionID).
					Return(session, nil)

				a.refreshSession.EXPECT().
					Update(mock.Anything, session).
					Return(errors.New("db unavailable"))
			},
			expectErr: ucerrs.ErrRevokeRefreshSessionDB,
		},
		{
			name: "Failure - account role lookup db error",
			input: dto.RefreshSessionInput{
				RefreshToken: rawToken,
				IP:           ip,
				UserAgent:    ua,
			},
			mockBehaviour: func(a adapter, accountID, sessionID uuid.UUID, rawToken string) {
				session := newTestSession(t, sessionID, accountID, utils.HashToken(rawToken), ip, ua)

				a.tokenGenerator.EXPECT().
					ValidateRefreshToken(mock.Anything, rawToken).
					Return(accountID, sessionID, nil)

				a.refreshSession.EXPECT().
					GetByID(mock.Anything, sessionID).
					Return(session, nil)

				a.refreshSession.EXPECT().
					Update(mock.Anything, session).
					Return(nil)

				a.accountRole.EXPECT().
					Get(mock.Anything, accountID).
					Return(nil, errors.New("db unavailable"))
			},
			expectErr: ucerrs.ErrGetAccountRoleDB,
			// NOTE: this case also demonstrates bug #2 from the review — by the
			// time this fails, the old session has already been revoked and
			// Update() succeeded, so the user is logged out with no new tokens.
		},
		{
			name: "Failure - token pair generation error",
			input: dto.RefreshSessionInput{
				RefreshToken: rawToken,
				IP:           ip,
				UserAgent:    ua,
			},
			mockBehaviour: func(a adapter, accountID, sessionID uuid.UUID, rawToken string) {
				session := newTestSession(t, sessionID, accountID, utils.HashToken(rawToken), ip, ua)

				a.tokenGenerator.EXPECT().
					ValidateRefreshToken(mock.Anything, rawToken).
					Return(accountID, sessionID, nil)

				a.refreshSession.EXPECT().
					GetByID(mock.Anything, sessionID).
					Return(session, nil)

				a.refreshSession.EXPECT().
					Update(mock.Anything, session).
					Return(nil)

				a.accountRole.EXPECT().
					Get(mock.Anything, accountID).
					Return(model.RestoreAccountRole(accountID, model.RoleUser), nil)

				a.tokenGenerator.EXPECT().
					GeneratePair(mock.Anything, accountID, model.RoleUser.String(), mock.AnythingOfType("uuid.UUID")).
					Return(&port.TokensPair{}, errors.New("signing failure"))
			},
			expectErr: ucerrs.ErrGenerateTokensPair,
		},
		{
			name: "Failure - create new session db error",
			input: dto.RefreshSessionInput{
				RefreshToken: rawToken,
				IP:           ip,
				UserAgent:    ua,
			},
			mockBehaviour: func(a adapter, accountID, sessionID uuid.UUID, rawToken string) {
				session := newTestSession(t, sessionID, accountID, utils.HashToken(rawToken), ip, ua)

				a.tokenGenerator.EXPECT().
					ValidateRefreshToken(mock.Anything, rawToken).
					Return(accountID, sessionID, nil)

				a.refreshSession.EXPECT().
					GetByID(mock.Anything, sessionID).
					Return(session, nil)

				a.refreshSession.EXPECT().
					Update(mock.Anything, session).
					Return(nil)

				a.accountRole.EXPECT().
					Get(mock.Anything, accountID).
					Return(model.RestoreAccountRole(accountID, model.RoleUser), nil)

				a.tokenGenerator.EXPECT().
					GeneratePair(mock.Anything, accountID, model.RoleUser.String(), mock.AnythingOfType("uuid.UUID")).
					Return(&port.TokensPair{Access: "new-access", Refresh: "new-refresh"}, nil)

				a.refreshSession.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*model.RefreshSession")).
					Return(errors.New("db unavailable"))
			},
			expectErr: ucerrs.ErrCreateRefreshSessionDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountID := uuid.New()
			sessionID := uuid.New()

			// Mocks
			accountRoleRepo := mocks.NewMockAccountRoleRepository(t)
			refreshSessionRepo := mocks.NewMockRefreshSessionRepository(t)
			tokenGenerator := mocks.NewMockTokenGenerator(t)
			tt.mockBehaviour(adapter{
				accountRole:    accountRoleRepo,
				refreshSession: refreshSessionRepo,
				tokenGenerator: tokenGenerator,
			}, accountID, sessionID, rawToken)

			// UC
			uc := usecase.NewRefreshSessionUC(
				accountRoleRepo, refreshSessionRepo, tokenGenerator, refreshSessionTTL,
			)

			// Call method
			out, err := uc.Execute(context.Background(), tt.input)

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, out.AccessToken)
				assert.NotEmpty(t, out.RefreshToken)
			}
		})
	}
}
