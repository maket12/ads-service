package dto

type SearchAdsInput struct {
	Text      string
	Category  *string
	PriceFrom *int64
	PriceTo   *int64
	Limit     int32
	Offset    int32
	SortBy    string
}

type SearchAdsOutput struct {
	Items []AdIndexDTO
	Total int64
}
