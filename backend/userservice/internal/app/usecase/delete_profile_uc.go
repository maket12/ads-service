package usecase

import (
	"context"

	"github.com/maket12/ads-service/userservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/userservice/internal/app/errs"
	"github.com/maket12/ads-service/userservice/internal/domain/port"
)

type DeleteProfileUC struct {
	profile port.ProfileRepository
}

func NewDeleteProfileUC(profile port.ProfileRepository) *DeleteProfileUC {
	return &DeleteProfileUC{profile: profile}
}

func (uc *DeleteProfileUC) Execute(ctx context.Context, in dto.DeleteProfileInput) (dto.DeleteProfileOutput, error) {
	if err := uc.profile.Delete(ctx, in.AccountID); err != nil {
		return dto.DeleteProfileOutput{}, ucerrs.Wrap(
			ucerrs.ErrDeleteProfileDB, err,
		)
	}
	return dto.DeleteProfileOutput{Deleted: true}, nil
}
