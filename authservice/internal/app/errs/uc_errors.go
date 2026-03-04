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
================ Adapter failures ================
*/
var (
	ErrHashPassword         = errors.New("failed to hash password")
	ErrGenerateAccessToken  = errors.New("failed to generate access token")
	ErrGenerateRefreshToken = errors.New("failed to generate refresh token")
	ErrPublishEvent         = errors.New("failed to publish event")
)

/*
================ Database failures ================
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
)
