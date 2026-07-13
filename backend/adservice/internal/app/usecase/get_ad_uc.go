package usecase

import (
	"context"
	"errors"

	"github.com/maket12/ads-service/backend/adservice/internal/app/dto"
	"github.com/maket12/ads-service/backend/adservice/internal/app/errs"
	port2 "github.com/maket12/ads-service/backend/adservice/internal/domain/port"
	pkgerrs "github.com/maket12/ads-service/pkg/errs"
)

type GetAdUC struct {
	ad    port2.AdRepository
	media port2.MediaRepository
}

func NewGetAdUC(
	ad port2.AdRepository, media port2.MediaRepository,
) *GetAdUC {
	return &GetAdUC{
		ad:    ad,
		media: media,
	}
}

func (uc *GetAdUC) Execute(ctx context.Context, in dto.GetAdInput) (dto.GetAdOutput, error) {
	// Get from db
	ad, err := uc.ad.Get(ctx, in.AdID)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.GetAdOutput{}, errs.ErrInvalidAdID
		}
		return dto.GetAdOutput{}, errs.Wrap(
			errs.ErrGetAdDB, err,
		)
	}

	// Check if current user can see this ad
	if !ad.IsPublished() {
		if ad.SellerID() != in.SellerID {
			return dto.GetAdOutput{}, errs.ErrAccessDenied
		}
	}

	// Get images from db
	images, err := uc.media.Get(ctx, ad.ID())
	if err != nil {
		return dto.GetAdOutput{}, errs.Wrap(
			errs.ErrGetImagesDB, err,
		)
	}

	// Add images into rich model
	err = ad.Update(nil, nil, nil, images)
	if err != nil {
		return dto.GetAdOutput{}, errs.Wrap(
			errs.ErrInvalidInput, err,
		)
	}

	// Response
	return dto.GetAdOutput{
		AdID:        ad.ID(),
		SellerID:    ad.SellerID(),
		Title:       ad.Title(),
		Description: ad.Description(),
		Price:       ad.Price(),
		Status:      string(ad.Status()),
		Images:      ad.Images(),
		CreatedAt:   ad.CreatedAt(),
		UpdatedAt:   ad.UpdatedAt(),
	}, nil
}
