package usecase

import (
	"context"

	"github.com/maket12/ads-service/adservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/adservice/internal/app/errs"
	"github.com/maket12/ads-service/adservice/internal/app/mapper"
	"github.com/maket12/ads-service/adservice/internal/domain/port"
)

type ListAdsUC struct{ ad port.AdRepository }

func NewListAdsUC(ad port.AdRepository) *ListAdsUC {
	return &ListAdsUC{ad: ad}
}

func (uc *ListAdsUC) Execute(ctx context.Context, in dto.ListAdsInput) (dto.ListAdsOutput, error) {
	ads, err := uc.ad.ListAds(ctx, in.Limit, in.Offset)
	if err != nil {
		return dto.ListAdsOutput{}, ucerrs.Wrap(ucerrs.ErrListAdsDB, err)
	}
	return mapper.MapDomainToListAdsOut(ads), nil
}
