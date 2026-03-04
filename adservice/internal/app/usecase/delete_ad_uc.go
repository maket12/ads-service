package usecase

import (
	"ads/adservice/internal/app/dto"
	"ads/adservice/internal/app/uc_errors"
	"ads/adservice/internal/domain/port"
	"ads/pkg/errs"
	"context"
	"errors"
)

type DeleteAdUC struct {
	ad    port.AdRepository
	media port.MediaRepository
}

func NewDeleteAdUC(ad port.AdRepository, media port.MediaRepository) *DeleteAdUC {
	return &DeleteAdUC{ad: ad, media: media}
}

func (uc *DeleteAdUC) Execute(ctx context.Context, in dto.DeleteAdInput) (dto.DeleteAdOutput, error) {
	// Get from db
	ad, err := uc.ad.Get(ctx, in.AdID)
	if err != nil {
		if errors.Is(err, errs.ErrObjectNotFound) {
			return dto.DeleteAdOutput{Success: false}, uc_errors.ErrInvalidAdID
		}
		return dto.DeleteAdOutput{Success: false}, uc_errors.Wrap(
			uc_errors.ErrGetAdDB, err,
		)
	}

	// Check if current user can delete this ad
	if ad.SellerID() != in.SellerID {
		return dto.DeleteAdOutput{Success: false}, uc_errors.ErrAccessDenied
	}

	// Scenario №1: Delete status from database (if not published yet)
	if ad.IsOnModeration() {
		err = uc.ad.Delete(ctx, ad.ID())
		if err != nil {
			return dto.DeleteAdOutput{Success: false}, uc_errors.Wrap(
				uc_errors.ErrDeleteAdDB, err,
			)
		}

		err = uc.media.Delete(ctx, ad.ID())
		if err != nil {
			return dto.DeleteAdOutput{Success: false}, uc_errors.Wrap(
				uc_errors.ErrDeleteImagesDB, err,
			)
		}
	} else {
		// Scenario №2: Update status (deleted)
		err = ad.Delete()
		if err != nil {
			return dto.DeleteAdOutput{Success: false}, uc_errors.ErrCannotDelete
		}

		err = uc.ad.UpdateStatus(ctx, ad)
		if err != nil {
			return dto.DeleteAdOutput{Success: false}, uc_errors.Wrap(
				uc_errors.ErrUpdateAdStatusDB, err,
			)
		}
	}

	// Response
	return dto.DeleteAdOutput{Success: true}, nil
}
