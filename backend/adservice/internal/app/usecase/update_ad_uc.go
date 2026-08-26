package usecase

import (
	"context"
	"errors"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"github.com/maket12/ads-service/backend/adservice/v2/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/backend/adservice/v2/internal/app/errs"
	"github.com/maket12/ads-service/backend/adservice/v2/internal/domain/model"
	"github.com/maket12/ads-service/backend/adservice/v2/internal/domain/port"
	pkgerrs "github.com/maket12/ads-service/backend/adservice/v2/pkg/errs"
)

type UpdateAdUC struct {
	trManager trm.Manager
	ad        port.AdRepository
	media     port.MediaRepository
}

func NewUpdateAdUC(
	trManager trm.Manager,
	ad port.AdRepository,
	media port.MediaRepository,
) *UpdateAdUC {
	return &UpdateAdUC{
		trManager: trManager,
		ad:        ad,
		media:     media,
	}
}

func (uc *UpdateAdUC) Execute(ctx context.Context, in dto.UpdateAdInput) (dto.UpdateAdOutput, error) {
	// Get from db
	ad, err := uc.ad.Get(ctx, in.AdID)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.UpdateAdOutput{Success: false}, ucerrs.ErrAdNotFound
		}
		return dto.UpdateAdOutput{Success: false}, ucerrs.Wrap(
			ucerrs.ErrGetAdDB, err,
		)
	}

	// Check if current user can update this ad
	if ad.SellerID() != in.SellerID {
		return dto.UpdateAdOutput{Success: false}, ucerrs.ErrAccessDenied
	}

	// Update ad
	err = ad.Update(in.Title, in.Description, in.Price, in.Category, in.Images)
	if err != nil {
		if errors.Is(err, model.ErrAdCantBeUpdated) {
			return dto.UpdateAdOutput{}, ucerrs.ErrCannotUpdate
		}
		return dto.UpdateAdOutput{Success: false}, ucerrs.Wrap(
			ucerrs.ErrInvalidInput, err,
		)
	}

	// Update in databases
	if err = uc.trManager.Do(ctx, func(txCtx context.Context) error {
		updErr := uc.ad.Update(txCtx, ad)
		if updErr != nil {
			return ucerrs.Wrap(ucerrs.ErrUpdateAdDB, updErr)
		}

		if updErr = uc.media.Save(txCtx, ad.ID(), ad.Images()); updErr != nil {
			return ucerrs.Wrap(ucerrs.ErrSaveImagesDB, updErr)
		}

		return nil
	}); err != nil {
		return dto.UpdateAdOutput{Success: false}, err
	}

	// Response
	return dto.UpdateAdOutput{Success: true}, nil
}
