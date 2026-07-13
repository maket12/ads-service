package dto

type RefreshSessionInput struct {
	RefreshToken string
	IP           *string
	UserAgent    *string
}

type RefreshSessionOutput struct {
	AccessToken  string
	RefreshToken string
}
