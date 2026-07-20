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
	account          port.AccountRepository
	accountRole      port.AccountRoleRepository
	accountPublisher port.AccountPublisher
}

func NewDeleteAccountUC(
	account port.AccountRepository,
	accountRole port.AccountRoleRepository,
	accountPublisher port.AccountPublisher,
) *DeleteAccountUC {
	return &DeleteAccountUC{
		account:          account,
		accountRole:      accountRole,
		accountPublisher: accountPublisher,
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

	// Send an event to rabbitmq (delete profile)
	err = uc.accountPublisher.PublishAccountDelete(ctx, account.ID())
	if err != nil {
		return dto.DeleteAccountOutput{}, ucerrs.Wrap(
			ucerrs.ErrPublishEvent, err,
		)
	}

	return dto.DeleteAccountOutput{Deleted: true}, nil
}
