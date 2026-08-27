package errs

import "errors"

/*
================ Validation failures ================
*/
var (
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrInvalidRole         = errors.New("account cannot be assigned to this role")
	ErrCannotLogin         = errors.New("account is either blocked or not exists")
	ErrCannotLogout        = errors.New("session is already expired or revoked")
	ErrCannotBlock         = errors.New("account cannot be blocked due to its status or insufficient permissions")
	ErrCannotDelete        = errors.New("account cannot be deleted due to its status or insufficient permissions")
	ErrCannotVerify        = errors.New("verification token has been expired")
	ErrInvalidAccessToken  = errors.New("access token is invalid")
	ErrInvalidRefreshToken = errors.New("refresh token is invalid or not found")
	ErrCannotRevoke        = errors.New("refresh token has been already rotated or invalid")

	ErrInvalidInput = errors.New("invalid input") // for rich models
)

/*
================ Publisher and Email Sender failures ================
*/
var (
	ErrSendVerificationEmail = errors.New("failed to send verification email")
	ErrPublishEvent          = errors.New("failed to publish event")
)

/*
================ Repository failures ================
*/
var (
	ErrIndexAdES         = errors.New("failed to index ad using es")
	ErrDeleteAdIndexES   = errors.New("failed to delete ad index using es")
	ErrSearchAdIndexesES = errors.New("failed to search for ad indexes using es")

	ErrAdIndexNotFound = errors.New("ad index not found")
)
