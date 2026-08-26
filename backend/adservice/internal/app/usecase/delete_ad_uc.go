package usecase

import (
	"context"
	"errors"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"github.com/maket12/ads-service/backend/adservice/v2/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/backend/adservice/v2/internal/app/errs"
	"github.com/maket12/ads-service/backend/adservice/v2/internal/domain/port"
	pkgerrs "github.com/maket12/ads-service/backend/adservice/v2/pkg/errs"
)

type DeleteAdUC struct {
	trManager trm.Manager
	ad        port.AdRepository
	media     port.MediaRepository
}

func NewDeleteAdUC(
	trManager trm.Manager,
	ad port.AdRepository,
	media port.MediaRepository,
) *DeleteAdUC {
	return &DeleteAdUC{
		trManager: trManager,
		ad:        ad,
		media:     media,
	}
}

func (uc *DeleteAdUC) Execute(ctx context.Context, in dto.DeleteAdInput) (dto.DeleteAdOutput, error) {
	// Get from db
	ad, err := uc.ad.Get(ctx, in.AdID)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.DeleteAdOutput{Success: false}, ucerrs.ErrAdNotFound
		}
		return dto.DeleteAdOutput{Success: false}, ucerrs.Wrap(
			ucerrs.ErrGetAdDB, err,
		)
	}

	// Check if current user can delete this ad
	if ad.SellerID() != in.SellerID {
		return dto.DeleteAdOutput{Success: false}, ucerrs.ErrAccessDenied
	}

	// Check if the ad has been already deleted
	if ad.IsDeleted() {
		return dto.DeleteAdOutput{Success: false}, ucerrs.ErrCannotDelete
	}

	// Scenario №1: Delete status from database (if not published yet)
	if ad.IsOnModeration() {
		if err = uc.trManager.Do(ctx, func(txCtx context.Context) error {
			delErr := ad.Delete()
			if delErr != nil {
				return ucerrs.ErrCannotDelete
			}

			if delErr = uc.ad.Delete(txCtx, ad.ID()); delErr != nil {
				return ucerrs.Wrap(ucerrs.ErrDeleteAdDB, delErr)
			}

			if delErr = uc.media.Delete(txCtx, ad.ID()); delErr != nil {
				return ucerrs.Wrap(ucerrs.ErrDeleteImagesDB, delErr)
			}

			return nil
		}); err != nil {
			return dto.DeleteAdOutput{Success: false}, err
		}
	} else {
		// Scenario №2: Update status (deleted)
		err = ad.Delete()
		if err != nil {
			return dto.DeleteAdOutput{Success: false}, ucerrs.ErrCannotDelete
		}

		if err = uc.ad.Update(ctx, ad); err != nil {
			return dto.DeleteAdOutput{Success: false}, ucerrs.Wrap(
				ucerrs.ErrUpdateAdDB, err,
			)
		}

		if err = uc.media.Delete(ctx, ad.ID()); err != nil {
			return dto.DeleteAdOutput{Success: false}, ucerrs.Wrap(
				ucerrs.ErrDeleteImagesDB, err,
			)
		}
	}

	// Response
	return dto.DeleteAdOutput{Success: true}, nil
}
