package usecase

import (
	"context"
	"errors"

	"github.com/maket12/ads-service/adservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/adservice/internal/app/errs"
	"github.com/maket12/ads-service/adservice/internal/domain/port"
	pkgerrs "github.com/maket12/ads-service/pkg/errs"
)

type UpdateAdUC struct {
	ad    port.AdRepository
	media port.MediaRepository
}

func NewUpdateAdUC(
	ad port.AdRepository, media port.MediaRepository,
) *UpdateAdUC {
	return &UpdateAdUC{
		ad:    ad,
		media: media,
	}
}

func (uc *UpdateAdUC) Execute(ctx context.Context, in dto.UpdateAdInput) (dto.UpdateAdOutput, error) {
	// Get from db
	ad, err := uc.ad.Get(ctx, in.AdID)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.UpdateAdOutput{Success: false}, ucerrs.ErrInvalidAdID
		}
		return dto.UpdateAdOutput{Success: false}, ucerrs.Wrap(
			ucerrs.ErrGetAdDB, err,
		)
	}

	// Check if current user can update this ad
	if ad.SellerID() != in.SellerID {
		return dto.UpdateAdOutput{Success: false}, ucerrs.ErrAccessDenied
	}

	// Update
	err = ad.Update(in.Title, in.Description, in.Price, in.Images)
	if err != nil {
		return dto.UpdateAdOutput{Success: false}, ucerrs.Wrap(
			ucerrs.ErrInvalidInput, err,
		)
	}

	// Update in db
	err = uc.ad.Update(ctx, ad)
	if err != nil {
		return dto.UpdateAdOutput{Success: false}, ucerrs.Wrap(
			ucerrs.ErrUpdateAdDB, err,
		)
	}

	// Update images in db
	err = uc.media.Save(ctx, ad.ID(), ad.Images())
	if err != nil {
		return dto.UpdateAdOutput{Success: false}, ucerrs.Wrap(
			ucerrs.ErrSaveImagesDB, err,
		)
	}

	// Response
	return dto.UpdateAdOutput{Success: true}, nil
}
