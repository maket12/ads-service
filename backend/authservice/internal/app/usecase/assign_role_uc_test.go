package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/maket12/ads-service/authservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/authservice/internal/app/errs"
	"github.com/maket12/ads-service/authservice/internal/app/usecase"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	"github.com/maket12/ads-service/authservice/internal/domain/port/mocks"
	pkgerrs "github.com/maket12/ads-service/authservice/pkg/errs"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAssignRoleUC_Execute(t *testing.T) {
	type adapter struct {
		accountRole    *mocks.MockAccountRoleRepository
		refreshSession *mocks.MockRefreshSessionRepository
	}

	type testCase struct {
		name          string
		input         dto.AssignRoleInput
		mockBehaviour func(a adapter, accountID uuid.UUID)
		expectErr     error
		expectAssign  bool
	}

	var tests = []testCase{
		{
			name: "Success",
			input: dto.AssignRoleInput{
				Role: model.RoleUser.String(),
			},
			mockBehaviour: func(a adapter, accountID uuid.UUID) {
				accRole := model.RestoreAccountRole(accountID, model.RoleAdmin)

				a.accountRole.EXPECT().
					Get(mock.Anything, accountID).
					Return(accRole, nil)

				a.accountRole.EXPECT().
					Update(mock.Anything, accRole).
					Return(nil)

				a.refreshSession.EXPECT().
					RevokeAllForAccount(mock.Anything, accountID, mock.Anything).
					Return(nil)
			},
			expectErr:    nil,
			expectAssign: true,
		},
		{
			name: "Failure - account role not found",
			input: dto.AssignRoleInput{
				Role: model.RoleAdmin.String(),
			},
			mockBehaviour: func(a adapter, accountID uuid.UUID) {
				a.accountRole.EXPECT().
					Get(mock.Anything, accountID).
					Return(nil, pkgerrs.ErrObjectNotFound)
			},
			expectErr:    ucerrs.ErrInvalidAccountID,
			expectAssign: false,
		},
		{
			name: "Failure - Get account role db error",
			input: dto.AssignRoleInput{
				Role: model.RoleAdmin.String(),
			},
			mockBehaviour: func(a adapter, accountID uuid.UUID) {
				a.accountRole.EXPECT().
					Get(mock.Anything, accountID).
					Return(nil, errors.New("db connection error"))
			},
			expectErr:    ucerrs.ErrGetAccountRoleDB,
			expectAssign: false,
		},
		{
			name: "Failure - Update account role db error",
			input: dto.AssignRoleInput{
				Role: model.RoleAdmin.String(),
			},
			mockBehaviour: func(a adapter, accountID uuid.UUID) {
				accRole := model.RestoreAccountRole(accountID, model.RoleUser)

				a.accountRole.EXPECT().
					Get(mock.Anything, accountID).
					Return(accRole, nil)

				a.accountRole.EXPECT().
					Update(mock.Anything, accRole).
					Return(errors.New("write error"))
			},
			expectErr:    ucerrs.ErrUpdateAccountRoleDB,
			expectAssign: false,
		},
		{
			name: "Failure - Revoke refresh sessions db error",
			input: dto.AssignRoleInput{
				Role: model.RoleAdmin.String(),
			},
			mockBehaviour: func(a adapter, accountID uuid.UUID) {
				accRole := model.RestoreAccountRole(accountID, model.RoleUser)

				a.accountRole.EXPECT().
					Get(mock.Anything, accountID).
					Return(accRole, nil)

				a.accountRole.EXPECT().
					Update(mock.Anything, accRole).
					Return(nil)

				a.refreshSession.EXPECT().
					RevokeAllForAccount(mock.Anything, accountID, mock.Anything).
					Return(errors.New("db error"))
			},
			expectErr:    ucerrs.ErrRevokeRefreshSessionDB,
			expectAssign: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountID := uuid.New()
			tt.input.AccountID = accountID

			accountRoleRepo := mocks.NewMockAccountRoleRepository(t)
			refreshSessionRepo := mocks.NewMockRefreshSessionRepository(t)

			tt.mockBehaviour(adapter{
				accountRole:    accountRoleRepo,
				refreshSession: refreshSessionRepo,
			}, accountID)

			uc := usecase.NewAssignRoleUC(accountRoleRepo, refreshSessionRepo)

			out, err := uc.Execute(context.Background(), tt.input)

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectAssign, out.Assigned)
		})
	}
}
