package model

import pkgerrs "github.com/maket12/ads-service/backend/authservice/pkg/errs"

type SortOption string

const (
	SortByRelevance SortOption = "relevance"
	SortByPriceAsc  SortOption = "price_asc"
	SortByPriceDesc SortOption = "price_desc"
	SortByDateDesc  SortOption = "date_desc"
)

func (o SortOption) String() string { return string(o) }

func (o SortOption) IsValid() bool {
	switch o {
	case SortByRelevance, SortByPriceAsc,
		SortByPriceDesc, SortByDateDesc:
		return true
	default:
		return false
	}
}

func NewSortOption(val string) (SortOption, error) {
	if val == "" {
		return SortByRelevance, nil
	}
	option := SortOption(val)
	if !option.IsValid() {
		return "", pkgerrs.NewValueInvalidError("sort_option")
	}
	return option, nil
}

const (
	defaultLimit int32 = 20
	maxLimit     int32 = 100
)

// ================ Rich model for Search Query ================

type SearchQuery struct {
	text      string
	category  *Category
	priceFrom *int64
	priceTo   *int64
	limit     int32
	offset    int32
	sortBy    SortOption
}

func NewSearchQuery(
	text string,
	rawCategory *string,
	priceFrom, priceTo *int64,
	limit, offset int32,
	sortBy string,
) (*SearchQuery, error) {
	var categoryPtr *Category
	if rawCategory != nil && *rawCategory != "" {
		cat, err := NewCategory(*rawCategory)
		if err != nil {
			return nil, err
		}
		categoryPtr = &cat
	}

	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if offset < 0 {
		offset = 0
	}

	if priceFrom != nil && *priceFrom < 0 {
		return nil, pkgerrs.NewValueInvalidError("price_from")
	}
	if priceTo != nil && *priceTo < 0 {
		return nil, pkgerrs.NewValueInvalidError("price_to")
	}
	if priceFrom != nil && priceTo != nil && *priceFrom > *priceTo {
		return nil, pkgerrs.NewValueInvalidError("price_range")
	}

	sortOption, err := NewSortOption(sortBy)
	if err != nil {
		return nil, err
	}

	return &SearchQuery{
		text:      text,
		category:  categoryPtr,
		priceFrom: priceFrom,
		priceTo:   priceTo,
		limit:     limit,
		offset:    offset,
		sortBy:    sortOption,
	}, nil
}

// ================ Read-Only ================

func (q *SearchQuery) Text() string        { return q.text }
func (q *SearchQuery) Category() *Category { return q.category }
func (q *SearchQuery) PriceFrom() *int64   { return q.priceFrom }
func (q *SearchQuery) PriceTo() *int64     { return q.priceTo }
func (q *SearchQuery) Limit() int32        { return q.limit }
func (q *SearchQuery) Offset() int32       { return q.offset }
func (q *SearchQuery) SortBy() SortOption  { return q.sortBy }
