package errs

import "errors"

/*
================ Validation failures ================
*/
var (
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrCannotLogin         = errors.New("account either is blocked or not exists")
	ErrInvalidAccountID    = errors.New("account id is invalid or account with this id not found")
	ErrCannotAssign        = errors.New("account can not be assigned to this role")
	ErrInvalidRefreshToken = errors.New("refresh token is invalid or not found")
	ErrCannotRevoke        = errors.New("refresh token has been already rotated or invalid")
	ErrInvalidAccessToken  = errors.New("access token is invalid")

	ErrInvalidInput = errors.New("invalid input") // for rich models
)

/*
================ Infrastructure failures ================
*/
var (
	ErrHashPassword       = errors.New("failed to hash password")
	ErrGenerateTokensPair = errors.New("failed to generate tokens pair")
)

/*
================ Publisher and Email Sender failures ================
*/
var (
	ErrSendVerificationEmail = errors.New("failed to send verification email")
	ErrPublishEvent          = errors.New("failed to publish event")
)

/*
================ Repositories failures ================
*/
var (
	ErrCreateAccountDB         = errors.New("failed to create account using db")
	ErrAccountAlreadyExists    = errors.New("account with given email already exists")
	ErrCreateAccountRoleDB     = errors.New("failed to create account role using db")
	ErrGetAccountByEmailDB     = errors.New("failed to get account by email using db")
	ErrGetAccountByIDDB        = errors.New("failed to get account by id using db")
	ErrUpdateAccountDB         = errors.New("failed to update account using db")
	ErrGetAccountRoleDB        = errors.New("failed to get account role using db")
	ErrUpdateAccountRoleDB     = errors.New("failed to update account role using db")
	ErrCreateRefreshSessionDB  = errors.New("failed to create refresh session using db")
	ErrGetRefreshSessionByIDDB = errors.New("failed to get refresh session by ID using db")
	ErrRevokeRefreshSessionDB  = errors.New("failed to revoke refresh session using db")
	ErrRevokeAllForAccountDB   = errors.New("failed to revoke all refresh session for account using db")

	ErrAccountNotFound = errors.New("account not found")
)

var (
	ErrSaveVerificationTokenDB   = errors.New("failed to save verification token using db")
	ErrGetVerificationTokenDB    = errors.New("failed to get verification token using db")
	ErrDeleteVerificationTokenDB = errors.New("failed to delete verification token using db")

	ErrVerificationTokenNotFound = errors.New("verification token not found")
)
