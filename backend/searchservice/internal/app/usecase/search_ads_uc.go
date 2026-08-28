package usecase

import (
	"context"

	"github.com/maket12/ads-service/backend/searchservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/backend/searchservice/internal/app/errs"
	"github.com/maket12/ads-service/backend/searchservice/internal/app/mapper"
	"github.com/maket12/ads-service/backend/searchservice/internal/domain/model"
	"github.com/maket12/ads-service/backend/searchservice/internal/domain/port"
)

type SearchAdsUC struct {
	adIndex port.AdIndexRepository
}

func NewSearchAdsUC(adIndex port.AdIndexRepository) *SearchAdsUC {
	return &SearchAdsUC{adIndex: adIndex}
}

func (uc *SearchAdsUC) Execute(ctx context.Context, in dto.SearchAdsInput) (dto.SearchAdsOutput, error) {
	searchQuery, err := model.NewSearchQuery(
		in.Text, in.Category,
		in.PriceFrom, in.PriceTo,
		in.Limit, in.Offset, in.SortBy,
	)
	if err != nil {
		return dto.SearchAdsOutput{}, ucerrs.Wrap(
			ucerrs.ErrInvalidInput, err,
		)
	}

	items, total, err := uc.adIndex.Search(ctx, searchQuery)
	if err != nil {
		return dto.SearchAdsOutput{}, ucerrs.Wrap(
			ucerrs.ErrSearchAdIndexesES, err,
		)
	}

	return dto.SearchAdsOutput{
		Items: mapper.MapDomainAdIndexesToDTO(items),
		Total: total,
	}, nil
}
