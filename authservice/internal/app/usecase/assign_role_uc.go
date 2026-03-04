package usecase

import (
	"context"
	"errors"
	"github.com/maket12/ads-service/authservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/authservice/internal/app/errs"
	"github.com/maket12/ads-service/authservice/internal/domain/port"
	pkgerrs "github.com/maket12/ads-service/pkg/errs"
)

type AssignRoleUC struct {
	accountRole port.AccountRoleRepository
}

func NewAssignRoleUC(accountRole port.AccountRoleRepository) *AssignRoleUC {
	return &AssignRoleUC{accountRole: accountRole}
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
	if err := accRole.Assign(in.Role); err != nil {
		return dto.AssignRoleOutput{Assign: false},
			ucerrs.ErrCannotAssign
	}

	// Update db
	if err := uc.accountRole.Update(ctx, accRole); err != nil {
		return dto.AssignRoleOutput{Assign: false},
			ucerrs.Wrap(ucerrs.ErrUpdateAccountRoleDB, err)
	}

	// Output
	return dto.AssignRoleOutput{Assign: true}, nil
}
