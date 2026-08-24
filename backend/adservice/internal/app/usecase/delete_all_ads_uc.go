package usecase

import (
	"context"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"github.com/maket12/ads-service/backend/adservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/backend/adservice/internal/app/errs"
	"github.com/maket12/ads-service/backend/adservice/internal/domain/port"
)

type DeleteAllAdsUC struct {
	trManager trm.Manager
	ad        port.AdRepository
	media     port.MediaRepository
}

func NewDeleteAllAdsUC(
	trManager trm.Manager,
	ad port.AdRepository,
	media port.MediaRepository,
) *DeleteAllAdsUC {
	return &DeleteAllAdsUC{
		trManager: trManager,
		ad:        ad,
		media:     media,
	}
}

func (uc *DeleteAllAdsUC) Execute(ctx context.Context, in dto.DeleteAllAdsInput) (dto.DeleteAllAdsOutput, error) {
	ads, getErr := uc.ad.ListSellerAds(ctx, in.SellerID)
	if getErr != nil {
		return dto.DeleteAllAdsOutput{Success: false}, ucerrs.Wrap(
			ucerrs.ErrListSellerAdsDB, getErr,
		)
	}

	// Delete all ads and their medias (if the ad hasn't been published yet)
	if err := uc.trManager.Do(ctx, func(txCtx context.Context) error {
		for _, ad := range ads {
			if ad.IsDeleted() {
				continue
			}
			// Scenario №1: Delete status from database (if not published yet)
			if ad.IsOnModeration() {
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
			} else {
				// Scenario №2: Update status (deleted)
				if delErr := ad.Delete(); delErr != nil {
					return ucerrs.ErrCannotDelete
				}
				if updErr := uc.ad.Update(txCtx, ad); updErr != nil {
					return ucerrs.Wrap(ucerrs.ErrUpdateAdDB, updErr)
				}
			}
		}

		return nil
	}); err != nil {
		return dto.DeleteAllAdsOutput{Success: false}, err
	}

	// Response
	return dto.DeleteAllAdsOutput{Success: true}, nil
}
