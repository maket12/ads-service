package usecase

import (
	"context"
	"errors"

	"github.com/maket12/ads-service/authservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/authservice/internal/app/errs"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	"github.com/maket12/ads-service/authservice/internal/domain/port"
	pkgerrs "github.com/maket12/ads-service/pkg/errs"
)

type RegisterUC struct {
	account          port.AccountRepository
	accountRole      port.AccountRoleRepository
	passwordHasher   port.PasswordHasher
	accountPublisher port.AccountPublisher
}

func NewRegisterUC(
	account port.AccountRepository,
	accountRole port.AccountRoleRepository,
	passwordHasher port.PasswordHasher,
	accountPublisher port.AccountPublisher,
) *RegisterUC {
	return &RegisterUC{
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
	if err := uc.account.Create(ctx, account); err != nil {
		if errors.Is(err, pkgerrs.ErrObjectAlreadyExists) {
			return dto.RegisterOutput{}, ucerrs.ErrAccountAlreadyExists
		}
		return dto.RegisterOutput{}, ucerrs.Wrap(
			ucerrs.ErrCreateAccountDB, err,
		)
	}
	if err := uc.accountRole.Create(ctx, accountRole); err != nil {
		return dto.RegisterOutput{}, ucerrs.Wrap(
			ucerrs.ErrCreateAccountRoleDB, err,
		)
	}

	// Send even to rabbitmq (create profile)
	if err := uc.accountPublisher.PublishAccountCreate(ctx, account.ID()); err != nil {
		return dto.RegisterOutput{},
			ucerrs.Wrap(ucerrs.ErrPublishEvent, err)
	}

	// Response
	return dto.RegisterOutput{AccountID: account.ID()}, nil
}
