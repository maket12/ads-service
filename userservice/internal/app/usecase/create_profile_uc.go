package usecase

import (
	"context"
	"errors"

	pkgerrs "github.com/maket12/ads-service/pkg/errs"
	"github.com/maket12/ads-service/userservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/userservice/internal/app/errs"
	"github.com/maket12/ads-service/userservice/internal/domain/model"
	"github.com/maket12/ads-service/userservice/internal/domain/port"
)

type CreateProfileUC struct {
	profile port.ProfileRepository
}

func NewCreateProfileUC(profile port.ProfileRepository) *CreateProfileUC {
	return &CreateProfileUC{profile: profile}
}

func (uc *CreateProfileUC) Execute(ctx context.Context, in dto.CreateProfileInput) error {
	// Create profile
	profile, err := model.NewProfile(in.AccountID)
	if err != nil {
		return ucerrs.ErrInvalidAccountID
	}

	if err := uc.profile.Create(ctx, profile); err != nil {
		if errors.Is(err, pkgerrs.ErrObjectAlreadyExists) {
			return nil
		}
		return ucerrs.Wrap(ucerrs.ErrCreateProfileDB, err)
	}

	return nil
}
