package dto

type LogoutInput struct {
	RefreshToken string
}

type LogoutOutput struct {
	Logout bool
}
