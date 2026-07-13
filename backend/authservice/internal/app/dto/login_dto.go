package dto

type LoginInput struct {
	Email     string
	Password  string
	IP        *string
	UserAgent *string
}

type LoginOutput struct {
	AccessToken  string
	RefreshToken string
}
