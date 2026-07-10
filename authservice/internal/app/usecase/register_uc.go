package usecase

import (
	"context"
	"errors"

	"github.com/maket12/ads-service/authservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/authservice/internal/app/errs"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	"github.com/maket12/ads-service/authservice/internal/domain/port"
	pkgerrs "github.com/maket12/ads-service/authservice/pkg/errs"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
)

type RegisterUC struct {
	trManager        trm.Manager
	account          port.AccountRepository
	accountRole      port.AccountRoleRepository
	passwordHasher   port.PasswordHasher
	accountPublisher port.AccountPublisher
}

func NewRegisterUC(
	trManager trm.Manager,
	account port.AccountRepository,
	accountRole port.AccountRoleRepository,
	passwordHasher port.PasswordHasher,
	accountPublisher port.AccountPublisher,
) *RegisterUC {
	return &RegisterUC{
		trManager:        trManager,
		account:          account,
		accountRole:      accountRole,
		passwordHasher:   passwordHasher,
		accountPublisher: accountPublisher,
	}
}

func (uc *RegisterUC) Execute(ctx context.Context, in dto.RegisterInput) (dto.RegisterOutput, error) {
	// Hashing the password
	hashedPassword, err := uc.passwordHasher.Hash(in.Password)
	if err != nil {
		return dto.RegisterOutput{}, ucerrs.Wrap(
			ucerrs.ErrHashPassword, err,
		)
	}

	// Creating rich-models with validation
	account, err := model.NewAccount(in.Email, hashedPassword)
	if err != nil {
		return dto.RegisterOutput{}, ucerrs.Wrap(
			ucerrs.ErrInvalidInput, err,
		)
	}
	accountRole, err := model.NewAccountRole(account.ID())
	if err != nil {
		return dto.RegisterOutput{}, ucerrs.Wrap(
			ucerrs.ErrInvalidInput, err,
		)
	}

	// Save all into database
	err = uc.trManager.Do(ctx, func(txCtx context.Context) error {
		createErr := uc.account.Create(ctx, account)
		if createErr != nil {
			if errors.Is(err, pkgerrs.ErrObjectAlreadyExists) {
				return ucerrs.ErrAccountAlreadyExists
			}
			return ucerrs.Wrap(ucerrs.ErrCreateAccountDB, err)
		}

		createErr = uc.accountRole.Create(ctx, accountRole)
		if createErr != nil {
			return ucerrs.Wrap(ucerrs.ErrCreateAccountRoleDB, err)
		}

		return nil
	})
	if err != nil {
		return dto.RegisterOutput{}, err
	}

	// Send an event to rabbitmq (create profile)
	if err = uc.accountPublisher.PublishAccountCreate(ctx, account.ID()); err != nil {
		return dto.RegisterOutput{}, ucerrs.Wrap(
			ucerrs.ErrPublishEvent, err,
		)
	}

	// Response
	return dto.RegisterOutput{AccountID: account.ID()}, nil
}
