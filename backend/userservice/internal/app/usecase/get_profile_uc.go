package usecase

import (
	"context"
	"errors"

	pkgerrs "github.com/maket12/ads-service/pkg/errs"
	"github.com/maket12/ads-service/userservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/userservice/internal/app/errs"
	"github.com/maket12/ads-service/userservice/internal/domain/port"
)

type GetProfileUC struct {
	profile port.ProfileRepository
}

func NewGetProfileUC(profile port.ProfileRepository) *GetProfileUC {
	return &GetProfileUC{profile: profile}
}

func (uc *GetProfileUC) Execute(ctx context.Context, in dto.GetProfileInput) (dto.GetProfileOutput, error) {
	profile, err := uc.profile.Get(ctx, in.AccountID)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.GetProfileOutput{}, ucerrs.ErrInvalidAccountID
		}
		return dto.GetProfileOutput{},
			ucerrs.Wrap(ucerrs.ErrGetProfileDB, err)
	}
	return dto.GetProfileOutput{
		AccountID: profile.AccountID(),
		FirstName: profile.FirstName(),
		LastName:  profile.LastName(),
		Phone:     profile.Phone(),
		AvatarURL: profile.AvatarURL(),
		Bio:       profile.Bio(),
		UpdatedAt: profile.UpdatedAt(),
	}, nil
}
