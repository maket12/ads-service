package usecase

import (
	"context"
	"errors"

	pkgerrs "github.com/maket12/ads-service/pkg/errs"
	"github.com/maket12/ads-service/userservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/userservice/internal/app/errs"
	"github.com/maket12/ads-service/userservice/internal/domain/port"
)

type UpdateProfileUC struct {
	profile        port.ProfileRepository
	phoneValidator port.PhoneValidator
}

func NewUpdateProfileUC(
	profile port.ProfileRepository,
	phoneValidator port.PhoneValidator,
) *UpdateProfileUC {
	return &UpdateProfileUC{
		profile:        profile,
		phoneValidator: phoneValidator,
	}
}

func (uc *UpdateProfileUC) Execute(ctx context.Context, in dto.UpdateProfileInput) (dto.UpdateProfileOutput, error) {
	// Get from db
	profile, err := uc.profile.Get(ctx, in.AccountID)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.UpdateProfileOutput{Success: false},
				ucerrs.ErrInvalidAccountID
		}
		return dto.UpdateProfileOutput{Success: false},
			ucerrs.Wrap(ucerrs.ErrGetProfileDB, err)
	}

	// Phone number validation
	var validatedPhone *string
	if in.Phone != nil {
		normPhone, err := uc.phoneValidator.Validate(ctx, *in.Phone)
		if err != nil {
			return dto.UpdateProfileOutput{Success: false},
				ucerrs.ErrInvalidPhoneNumber
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
		return dto.UpdateProfileOutput{Success: false}, ucerrs.ErrInvalidProfileData
	}

	if err := uc.profile.Update(ctx, profile); err != nil {
		return dto.UpdateProfileOutput{Success: false},
			ucerrs.Wrap(ucerrs.ErrUpdateProfileDB, err)
	}

	return dto.UpdateProfileOutput{Success: true}, nil
}
