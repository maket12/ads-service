package usecase

import (
	"context"
	"errors"
	"github.com/maket12/ads-service/authservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/authservice/internal/app/errs"
	"github.com/maket12/ads-service/authservice/internal/domain/port"
	pkgerrs "github.com/maket12/ads-service/pkg/errs"
)

type ValidateAccessTokenUC struct {
	account        port.AccountRepository
	tokenGenerator port.TokenGenerator
}

func NewValidateAccessTokenUC(
	account port.AccountRepository,
	tokenGenerator port.TokenGenerator,
) *ValidateAccessTokenUC {
	return &ValidateAccessTokenUC{
		account:        account,
		tokenGenerator: tokenGenerator,
	}
}

func (uc *ValidateAccessTokenUC) Execute(ctx context.Context, in dto.ValidateAccessTokenInput) (dto.ValidateAccessTokenOutput, error) {
	// Parse access token
	accountID, role, err := uc.tokenGenerator.ValidateAccessToken(
		ctx, in.AccessToken,
	)
	if err != nil {
		return dto.ValidateAccessTokenOutput{}, ucerrs.ErrInvalidAccessToken
	}

	// Get account and check if it is not active
	account, err := uc.account.GetByID(ctx, accountID)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.ValidateAccessTokenOutput{}, ucerrs.ErrInvalidAccessToken
		}
		return dto.ValidateAccessTokenOutput{}, ucerrs.Wrap(
			ucerrs.ErrGetAccountByIDDB, err,
		)
	}
	if !account.CanLogin() {
		return dto.ValidateAccessTokenOutput{}, ucerrs.ErrCannotLogin
	}

	// Output
	return dto.ValidateAccessTokenOutput{
		AccountID: accountID,
		Role:      role,
	}, nil
}
