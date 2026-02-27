package usecase

import (
	"ads/pkg/errs"
	"ads/userservice/internal/app/dto"
	"ads/userservice/internal/app/uc_errors"
	"ads/userservice/internal/domain/port"
	"context"
	"errors"
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
		if errors.Is(err, errs.ErrObjectNotFound) {
			return dto.UpdateProfileOutput{Success: false},
				uc_errors.ErrInvalidAccountID
		}
		return dto.UpdateProfileOutput{Success: false},
			uc_errors.Wrap(uc_errors.ErrGetProfileDB, err)
	}

	// Phone number validation
	var validatedPhone *string
	if in.Phone != nil {
		normPhone, err := uc.phoneValidator.Validate(ctx, *in.Phone)
		if err != nil {
			return dto.UpdateProfileOutput{Success: false},
				uc_errors.ErrInvalidPhoneNumber
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
		return dto.UpdateProfileOutput{Success: false}, uc_errors.ErrInvalidProfileData
	}

	if err := uc.profile.Update(ctx, profile); err != nil {
		return dto.UpdateProfileOutput{Success: false},
			uc_errors.Wrap(uc_errors.ErrUpdateProfileDB, err)
	}

	return dto.UpdateProfileOutput{Success: true}, nil
}
