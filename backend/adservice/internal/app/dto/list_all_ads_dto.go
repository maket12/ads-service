package dto

type ListAllAdsInput struct {
	Limit  int
	Offset int
}

type ListAllAdsOutput struct {
	Ads []Ad
}
