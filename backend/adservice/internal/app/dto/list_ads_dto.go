package dto

type ListAdsInput struct {
	Limit  int
	Offset int
}

type ListAdsOutput struct {
	Ads []Ad
}
