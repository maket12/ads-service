package dto

type VerifyEmailInput struct {
	Token string
}

type VerifyEmailOutput struct {
	Verified bool
}
