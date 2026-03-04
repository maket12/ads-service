package usecase

import (
	"ads/adservice/internal/app/dto"
	"ads/adservice/internal/app/uc_errors"
	"ads/adservice/internal/domain/port"
	"ads/pkg/errs"
	"context"
	"errors"
)

type GetAdUC struct {
	ad    port.AdRepository
	media port.MediaRepository
}

func NewGetAdUC(
	ad port.AdRepository, media port.MediaRepository,
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
		if errors.Is(err, errs.ErrObjectNotFound) {
			return dto.GetAdOutput{}, uc_errors.ErrInvalidAdID
		}
		return dto.GetAdOutput{}, uc_errors.Wrap(
			uc_errors.ErrGetAdDB, err,
		)
	}

	// Check if current user can see this ad
	if !ad.IsPublished() {
		if ad.SellerID() != in.SellerID {
			return dto.GetAdOutput{}, uc_errors.ErrAccessDenied
		}
	}

	// Get images from db
	images, err := uc.media.Get(ctx, ad.ID())
	if err != nil {
		return dto.GetAdOutput{}, uc_errors.Wrap(
			uc_errors.ErrGetImagesDB, err,
		)
	}

	// Add images into rich model
	err = ad.Update(nil, nil, nil, images)
	if err != nil {
		return dto.GetAdOutput{}, uc_errors.Wrap(
			uc_errors.ErrInvalidInput, err,
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
