package usecase

import (
	"context"

	"github.com/maket12/ads-service/backend/adservice/v2/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/backend/adservice/v2/internal/app/errs"
	"github.com/maket12/ads-service/backend/adservice/v2/internal/app/mapper"
	"github.com/maket12/ads-service/backend/adservice/v2/internal/domain/port"
)

type ListAllAdsUC struct{ ad port.AdRepository }

func NewListAllAdsUC(ad port.AdRepository) *ListAllAdsUC {
	return &ListAllAdsUC{ad: ad}
}

func (uc *ListAllAdsUC) Execute(ctx context.Context, in dto.ListAllAdsInput) (dto.ListAllAdsOutput, error) {
	ads, err := uc.ad.ListAds(ctx, in.Limit, in.Offset)
	if err != nil {
		return dto.ListAllAdsOutput{}, ucerrs.Wrap(ucerrs.ErrListAdsDB, err)
	}
	return mapper.MapDomainToListAdsOut(ads), nil
}
