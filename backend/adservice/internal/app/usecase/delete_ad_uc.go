package usecase

import (
	"context"
	"errors"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"github.com/maket12/ads-service/adservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/adservice/internal/app/errs"
	"github.com/maket12/ads-service/adservice/internal/domain/port"
	pkgerrs "github.com/maket12/ads-service/adservice/pkg/errs"
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

	// Scenario №1: Delete status from database (if not published yet)
	if ad.IsOnModeration() {
		if err = uc.trManager.Do(ctx, func(txCtx context.Context) error {
			delErr := uc.ad.Delete(ctx, ad.ID())
			if delErr != nil {
				return ucerrs.Wrap(ucerrs.ErrDeleteAdDB, delErr)
			}

			delErr = uc.media.Delete(ctx, ad.ID())
			if err != nil {
				return ucerrs.Wrap(ucerrs.ErrDeleteImagesDB, err)
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

		err = uc.ad.Update(ctx, ad)
		if err != nil {
			return dto.DeleteAdOutput{Success: false}, ucerrs.Wrap(
				ucerrs.ErrUpdateAdStatusDB, err,
			)
		}
	}

	// Response
	return dto.DeleteAdOutput{Success: true}, nil
}
