package usecase

import (
	"context"
	"errors"

	"github.com/maket12/ads-service/adservice/internal/app/dto"
	"github.com/maket12/ads-service/adservice/internal/app/errs"
	port2 "github.com/maket12/ads-service/adservice/internal/domain/port"
	pkgerrs "github.com/maket12/ads-service/adservice/pkg/errs"
)

type UpdateAdUC struct {
	ad    port2.AdRepository
	media port2.MediaRepository
}

func NewUpdateAdUC(
	ad port2.AdRepository, media port2.MediaRepository,
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
			return dto.UpdateAdOutput{Success: false}, errs.ErrInvalidAdID
		}
		return dto.UpdateAdOutput{Success: false}, errs.Wrap(
			errs.ErrGetAdDB, err,
		)
	}

	// Check if current user can update this ad
	if ad.SellerID() != in.SellerID {
		return dto.UpdateAdOutput{Success: false}, errs.ErrAccessDenied
	}

	// Update
	err = ad.Update(in.Title, in.Description, in.Price, in.Images)
	if err != nil {
		return dto.UpdateAdOutput{Success: false}, errs.Wrap(
			errs.ErrInvalidInput, err,
		)
	}

	// Update in db
	err = uc.ad.Update(ctx, ad)
	if err != nil {
		return dto.UpdateAdOutput{Success: false}, errs.Wrap(
			errs.ErrUpdateAdDB, err,
		)
	}

	// Update images in db
	err = uc.media.Save(ctx, ad.ID(), ad.Images())
	if err != nil {
		return dto.UpdateAdOutput{Success: false}, errs.Wrap(
			errs.ErrSaveImagesDB, err,
		)
	}

	// Response
	return dto.UpdateAdOutput{Success: true}, nil
}
