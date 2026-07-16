package usecase

import (
	"context"
	"errors"

	"github.com/maket12/ads-service/authservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/authservice/internal/app/errs"
	"github.com/maket12/ads-service/authservice/internal/domain/port"
	pkgerrs "github.com/maket12/ads-service/authservice/pkg/errs"
)

type DeleteAccountUC struct {
	account     port.AccountRepository
	accountRole port.AccountRoleRepository
}

func NewDeleteAccountUC(
	account port.AccountRepository,
	accountRole port.AccountRoleRepository,
) *DeleteAccountUC {
	return &DeleteAccountUC{
		account:     account,
		accountRole: accountRole,
	}
}

func (uc *DeleteAccountUC) Execute(ctx context.Context, in dto.DeleteAccountInput) (dto.DeleteAccountOutput, error) {
	// Find an account
	account, err := uc.account.GetByID(ctx, in.AccountID)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.DeleteAccountOutput{}, ucerrs.ErrAccountNotFound
		}
		return dto.DeleteAccountOutput{}, ucerrs.Wrap(
			ucerrs.ErrGetAccountByIDDB, err,
		)
	}

	// Get the account role to avoid deleting admins
	accountRole, err := uc.accountRole.Get(ctx, account.ID())
	if err != nil {
		return dto.DeleteAccountOutput{}, ucerrs.Wrap(
			ucerrs.ErrGetAccountRoleDB, err,
		)
	}

	if accountRole.IsAdmin() {
		return dto.DeleteAccountOutput{}, ucerrs.ErrCannotDelete
	}

	// Delete it and save into the database
	if err = account.Delete(); err != nil {
		return dto.DeleteAccountOutput{}, ucerrs.ErrCannotDelete
	}

	if err = uc.account.Update(ctx, account); err != nil {
		return dto.DeleteAccountOutput{}, ucerrs.Wrap(
			ucerrs.ErrUpdateAccountDB, err,
		)
	}

	return dto.DeleteAccountOutput{Deleted: true}, nil
}
