package usecase

import (
	"context"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"github.com/maket12/ads-service/backend/adservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/backend/adservice/internal/app/errs"
	"github.com/maket12/ads-service/backend/adservice/internal/domain/model"
	"github.com/maket12/ads-service/backend/adservice/internal/domain/port"
)

type CreateAdUC struct {
	trManager trm.Manager
	ad        port.AdRepository
	media     port.MediaRepository
}

func NewCreateAdUC(
	trManager trm.Manager,
	ad port.AdRepository,
	media port.MediaRepository,
) *CreateAdUC {
	return &CreateAdUC{
		trManager: trManager,
		ad:        ad,
		media:     media,
	}
}

func (uc *CreateAdUC) Execute(ctx context.Context, in dto.CreateAdInput) (dto.CreateAdOutput, error) {
	// Create ad
	ad, err := model.NewAd(
		in.SellerID, in.Title, in.Description,
		in.Price, in.Category, in.Images,
	)
	if err != nil {
		return dto.CreateAdOutput{}, ucerrs.Wrap(
			ucerrs.ErrInvalidInput, err,
		)
	}

	// Save into database
	if err = uc.trManager.Do(ctx, func(txCtx context.Context) error {
		if createErr := uc.ad.Create(txCtx, ad); createErr != nil {
			return ucerrs.Wrap(ucerrs.ErrCreateAdDB, createErr)
		}

		if saveErr := uc.media.Save(txCtx, ad.ID(), ad.Images()); saveErr != nil {
			return ucerrs.Wrap(ucerrs.ErrSaveImagesDB, saveErr)
		}

		return nil
	}); err != nil {
		return dto.CreateAdOutput{}, err
	}

	// Response
	return dto.CreateAdOutput{AdID: ad.ID()}, nil
}
