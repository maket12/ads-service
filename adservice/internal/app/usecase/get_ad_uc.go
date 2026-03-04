package usecase

import (
	"context"
	"errors"
	"github.com/maket12/ads-service/adservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/adservice/internal/app/errs"
	"github.com/maket12/ads-service/adservice/internal/domain/port"
	pkgerrs "github.com/maket12/ads-service/pkg/errs"
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
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.GetAdOutput{}, ucerrs.ErrInvalidAdID
		}
		return dto.GetAdOutput{}, ucerrs.Wrap(
			ucerrs.ErrGetAdDB, err,
		)
	}

	// Check if current user can see this ad
	if !ad.IsPublished() {
		if ad.SellerID() != in.SellerID {
			return dto.GetAdOutput{}, ucerrs.ErrAccessDenied
		}
	}

	// Get images from db
	images, err := uc.media.Get(ctx, ad.ID())
	if err != nil {
		return dto.GetAdOutput{}, ucerrs.Wrap(
			ucerrs.ErrGetImagesDB, err,
		)
	}

	// Add images into rich model
	err = ad.Update(nil, nil, nil, images)
	if err != nil {
		return dto.GetAdOutput{}, ucerrs.Wrap(
			ucerrs.ErrInvalidInput, err,
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
