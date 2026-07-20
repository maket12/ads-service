package usecase

import (
	"context"

	"github.com/maket12/ads-service/adservice/internal/app/dto"
	"github.com/maket12/ads-service/adservice/internal/app/errs"
	"github.com/maket12/ads-service/adservice/internal/domain/model"
	port2 "github.com/maket12/ads-service/adservice/internal/domain/port"
)

type CreateAdUC struct {
	ad    port2.AdRepository
	media port2.MediaRepository
}

func NewCreateAdUC(
	ad port2.AdRepository, media port2.MediaRepository,
) *CreateAdUC {
	return &CreateAdUC{
		ad:    ad,
		media: media,
	}
}

func (uc *CreateAdUC) Execute(ctx context.Context, in dto.CreateAdInput) (dto.CreateAdOutput, error) {
	// Create ad
	ad, err := model.NewAd(
		in.SellerID, in.Title,
		in.Description, in.Price, in.Images,
	)
	if err != nil {
		return dto.CreateAdOutput{}, errs.Wrap(
			errs.ErrInvalidInput, err,
		)
	}

	// Save into database
	if err := uc.ad.Create(ctx, ad); err != nil {
		return dto.CreateAdOutput{}, errs.Wrap(
			errs.ErrCreateAdDB, err,
		)
	}

	// Save images into database
	if err := uc.media.Save(ctx, ad.ID(), ad.Images()); err != nil {
		return dto.CreateAdOutput{}, errs.Wrap(errs.ErrSaveImagesDB, err)
	}

	// Response
	return dto.CreateAdOutput{AdID: ad.ID()}, nil
}
