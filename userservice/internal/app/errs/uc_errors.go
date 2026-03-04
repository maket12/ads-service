package errs

import "errors"

/*
================ Validation failures ================
*/
var (
	ErrInvalidAccountID   = errors.New("account id is invalid or account with this id not exist")
	ErrInvalidProfileData = errors.New("invalid profile data")
	ErrInvalidPhoneNumber = errors.New("invalid phone number")
)

/*
================ Database failures ================
*/
var (
	ErrCreateProfileDB = errors.New("failed to create profile using db")
	ErrGetProfileDB    = errors.New("failed to get profile using db")
	ErrUpdateProfileDB = errors.New("failed to update profile using db")
)
