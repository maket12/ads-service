package usecase

import (
	"context"
	"errors"

	"github.com/maket12/ads-service/authservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/authservice/internal/app/errs"
	"github.com/maket12/ads-service/authservice/internal/domain/port"
	pkgerrs "github.com/maket12/ads-service/authservice/pkg/errs"
)

type BlockAccountUC struct {
	account     port.AccountRepository
	accountRole port.AccountRoleRepository
}

func NewBlockAccountUC(
	account port.AccountRepository,
	accountRole port.AccountRoleRepository,
) *BlockAccountUC {
	return &BlockAccountUC{
		account:     account,
		accountRole: accountRole,
	}
}

func (uc *BlockAccountUC) Execute(ctx context.Context, in dto.BlockAccountInput) (dto.BlockAccountOutput, error) {
	// Find an account
	account, err := uc.account.GetByID(ctx, in.AccountID)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.BlockAccountOutput{}, ucerrs.ErrAccountNotFound
		}
		return dto.BlockAccountOutput{}, ucerrs.Wrap(
			ucerrs.ErrGetAccountByIDDB, err,
		)
	}

	// Get the account role to avoid blocking admins
	accountRole, err := uc.accountRole.Get(ctx, account.ID())
	if err != nil {
		return dto.BlockAccountOutput{}, ucerrs.Wrap(
			ucerrs.ErrGetAccountRoleDB, err,
		)
	}

	if accountRole.IsAdmin() {
		return dto.BlockAccountOutput{}, ucerrs.ErrCannotBlock
	}

	// Block it and save into the database
	if err = account.Block(); err != nil {
		return dto.BlockAccountOutput{}, ucerrs.ErrCannotBlock
	}

	if err = uc.account.Update(ctx, account); err != nil {
		return dto.BlockAccountOutput{}, ucerrs.Wrap(
			ucerrs.ErrUpdateAccountDB, err,
		)
	}

	return dto.BlockAccountOutput{Blocked: true}, nil
}
