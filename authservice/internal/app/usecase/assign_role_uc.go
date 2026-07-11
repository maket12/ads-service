package usecase

import (
	"context"
	"errors"

	"github.com/maket12/ads-service/authservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/authservice/internal/app/errs"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	"github.com/maket12/ads-service/authservice/internal/domain/port"
	pkgerrs "github.com/maket12/ads-service/authservice/pkg/errs"
	"github.com/maket12/ads-service/pkg/utils"
)

type AssignRoleUC struct {
	accountRole    port.AccountRoleRepository
	refreshSession port.RefreshSessionRepository
}

func NewAssignRoleUC(
	accountRole port.AccountRoleRepository,
	refreshSession port.RefreshSessionRepository,
) *AssignRoleUC {
	return &AssignRoleUC{
		accountRole:    accountRole,
		refreshSession: refreshSession,
	}
}

func (uc *AssignRoleUC) Execute(ctx context.Context, in dto.AssignRoleInput) (dto.AssignRoleOutput, error) {
	// Get role
	accRole, err := uc.accountRole.Get(ctx, in.AccountID)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.AssignRoleOutput{Assign: false},
				ucerrs.ErrInvalidAccountID
		}
		return dto.AssignRoleOutput{Assign: false},
			ucerrs.Wrap(ucerrs.ErrGetAccountRoleDB, err)
	}

	// Assign
	if err = accRole.Assign(in.Role); err != nil {
		return dto.AssignRoleOutput{Assign: false},
			ucerrs.ErrCannotAssign
	}

	// Update db
	if err = uc.accountRole.Update(ctx, accRole); err != nil {
		return dto.AssignRoleOutput{Assign: false},
			ucerrs.Wrap(ucerrs.ErrUpdateAccountRoleDB, err)
	}

	// Revoke all refresh tokens for security
	err = uc.refreshSession.RevokeAllForAccount(
		ctx,
		in.AccountID,
		utils.VPtr(model.ReasonRoleChanged.String()),
	)
	if err != nil {
		return dto.AssignRoleOutput{Assign: false}, ucerrs.Wrap(
			ucerrs.ErrRevokeRefreshSessionDB, err,
		)
	}

	// Output
	return dto.AssignRoleOutput{Assign: true}, nil
}
