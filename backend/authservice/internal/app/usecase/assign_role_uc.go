package usecase

import (
	"context"
	"errors"

	"github.com/maket12/ads-service/authservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/authservice/internal/app/errs"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	"github.com/maket12/ads-service/authservice/internal/domain/port"
	pkgerrs "github.com/maket12/ads-service/authservice/pkg/errs"
	"github.com/maket12/ads-service/authservice/pkg/utils"
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
			return dto.AssignRoleOutput{}, ucerrs.ErrAccountNotFound
		}
		return dto.AssignRoleOutput{}, ucerrs.Wrap(
			ucerrs.ErrGetAccountRoleDB, err,
		)
	}

	// Assign the account and save it into database
	if err = accRole.Assign(in.Role); err != nil {
		return dto.AssignRoleOutput{}, ucerrs.ErrInvalidRole
	}

	if err = uc.accountRole.Update(ctx, accRole); err != nil {
		return dto.AssignRoleOutput{}, ucerrs.Wrap(
			ucerrs.ErrUpdateAccountRoleDB, err,
		)
	}

	// Revoke all refresh tokens for security
	err = uc.refreshSession.RevokeAllForAccount(ctx, in.AccountID,
		utils.VPtr(model.ReasonRoleChanged.String()),
	)
	if err != nil {
		return dto.AssignRoleOutput{}, ucerrs.Wrap(
			ucerrs.ErrRevokeRefreshSessionDB, err,
		)
	}

	// Output
	return dto.AssignRoleOutput{Assigned: true}, nil
}
