package usecase

import (
	"context"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"github.com/maket12/ads-service/adservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/adservice/internal/app/errs"
	"github.com/maket12/ads-service/adservice/internal/domain/port"
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
			delErr := ad.Delete()
			if delErr != nil {
				return ucerrs.Wrap(ucerrs.ErrInvalidInput, delErr)
			}
		}
		if delErr := uc.ad.DeleteAll(ctx, in.SellerID); err != nil {
			return ucerrs.Wrap(ucerrs.ErrDeleteAllAdsDB, err)
		}
	}); err != nil {
		return dto.DeleteAllAdsOutput{Success: false}, err
	}

	// Response
	return dto.DeleteAllAdsOutput{Success: true}, nil
}
