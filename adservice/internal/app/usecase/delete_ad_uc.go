package usecase

import (
	"context"
	"errors"

	"github.com/maket12/ads-service/adservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/adservice/internal/app/errs"
	"github.com/maket12/ads-service/adservice/internal/domain/port"
	pkgerrs "github.com/maket12/ads-service/pkg/errs"
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
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.DeleteAdOutput{Success: false}, ucerrs.ErrInvalidAdID
		}
		return dto.DeleteAdOutput{Success: false}, ucerrs.Wrap(
			ucerrs.ErrGetAdDB, err,
		)
	}

	// Check if current user can delete this ad
	if ad.SellerID() != in.SellerID {
		return dto.DeleteAdOutput{Success: false}, ucerrs.ErrAccessDenied
	}

	// Scenario №1: Delete status from database (if not published yet)
	if ad.IsOnModeration() {
		err = uc.ad.Delete(ctx, ad.ID())
		if err != nil {
			return dto.DeleteAdOutput{Success: false}, ucerrs.Wrap(
				ucerrs.ErrDeleteAdDB, err,
			)
		}

		err = uc.media.Delete(ctx, ad.ID())
		if err != nil {
			return dto.DeleteAdOutput{Success: false}, ucerrs.Wrap(
				ucerrs.ErrDeleteImagesDB, err,
			)
		}
	} else {
		// Scenario №2: Update status (deleted)
		err = ad.Delete()
		if err != nil {
			return dto.DeleteAdOutput{Success: false}, ucerrs.ErrCannotDelete
		}

		err = uc.ad.UpdateStatus(ctx, ad)
		if err != nil {
			return dto.DeleteAdOutput{Success: false}, ucerrs.Wrap(
				ucerrs.ErrUpdateAdStatusDB, err,
			)
		}
	}

	// Response
	return dto.DeleteAdOutput{Success: true}, nil
}
