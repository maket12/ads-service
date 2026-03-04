package usecase

import (
	"context"
	"github.com/maket12/ads-service/adservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/adservice/internal/app/errs"
	"github.com/maket12/ads-service/adservice/internal/domain/port"
)

type DeleteAllAdsUC struct {
	ad port.AdRepository
}

func NewDeleteAllAdsUC(ad port.AdRepository) *DeleteAllAdsUC {
	return &DeleteAllAdsUC{ad: ad}
}

func (uc *DeleteAllAdsUC) Execute(ctx context.Context, in dto.DeleteAllAdsInput) (dto.DeleteAllAdsOutput, error) {
	// Delete all ads
	if err := uc.ad.DeleteAll(ctx, in.SellerID); err != nil {
		return dto.DeleteAllAdsOutput{Success: false}, ucerrs.Wrap(
			ucerrs.ErrDeleteAllAdsDB, err,
		)
	}

	// Response
	return dto.DeleteAllAdsOutput{Success: true}, nil
}
