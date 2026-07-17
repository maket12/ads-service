package usecase

import (
	"context"
	"errors"

	"github.com/maket12/ads-service/userservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/userservice/internal/app/errs"
	"github.com/maket12/ads-service/userservice/internal/domain/port"
	pkgerrs "github.com/maket12/ads-service/userservice/pkg/errs"
)

type UpdateProfileUC struct {
	profile port.ProfileRepository
	phone   port.PhoneValidator
}

func NewUpdateProfileUC(
	profile port.ProfileRepository,
	phone port.PhoneValidator,
) *UpdateProfileUC {
	return &UpdateProfileUC{
		profile: profile,
		phone:   phone,
	}
}

func (uc *UpdateProfileUC) Execute(ctx context.Context, in dto.UpdateProfileInput) (dto.UpdateProfileOutput, error) {
	// Get from db
	profile, err := uc.profile.Get(ctx, in.AccountID)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.UpdateProfileOutput{}, ucerrs.ErrProfileNotFound
		}
		return dto.UpdateProfileOutput{}, ucerrs.Wrap(
			ucerrs.ErrGetProfileDB, err,
		)
	}

	// Phone number validation
	var validatedPhone *string
	if in.Phone != nil {
		normPhone, phoneErr := uc.phone.Validate(ctx, *in.Phone)
		if phoneErr != nil {
			return dto.UpdateProfileOutput{}, ucerrs.Wrap(
				ucerrs.ErrInvalidInput, pkgerrs.NewValueInvalidErrorWithReason(
					"phone", phoneErr,
				),
			)
		}
		validatedPhone = &normPhone
	}

	// Update
	err = profile.Update(
		in.FirstName,
		in.LastName,
		validatedPhone,
		in.AvatarURL,
		in.Bio,
	)
	if err != nil {
		return dto.UpdateProfileOutput{}, ucerrs.Wrap(
			ucerrs.ErrInvalidInput, err,
		)
	}

	if err = uc.profile.Update(ctx, profile); err != nil {
		return dto.UpdateProfileOutput{}, ucerrs.Wrap(
			ucerrs.ErrUpdateProfileDB, err,
		)
	}

	return dto.UpdateProfileOutput{Success: true}, nil
}
