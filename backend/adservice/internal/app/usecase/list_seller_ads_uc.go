package usecase

import (
	"context"

	"github.com/maket12/ads-service/adservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/adservice/internal/app/errs"
	"github.com/maket12/ads-service/adservice/internal/app/mapper"
	"github.com/maket12/ads-service/adservice/internal/domain/port"
)

type ListSellerAdsUC struct{ ad port.AdRepository }

func NewListSellerAdsUC(ad port.AdRepository) *ListSellerAdsUC {
	return &ListSellerAdsUC{ad: ad}
}

func (uc *ListSellerAdsUC) Execute(ctx context.Context, in dto.ListSellerAdsInput) (dto.ListSellerAdsOutput, error) {
	ads, err := uc.ad.ListSellerAds(ctx, in.SellerID)
	if err != nil {
		return dto.ListSellerAdsOutput{}, ucerrs.Wrap(
			ucerrs.ErrListSellerAdsDB, err,
		)
	}
	return mapper.MapDomainToListSellerAdsOut(ads), nil
}
