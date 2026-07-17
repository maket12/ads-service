package errs

import "errors"

/*
================ Validation failures ================
*/
var ErrInvalidInput = errors.New("invalid input") // for rich models

/*
================ Database failures ================
*/
var (
	ErrCreateProfileDB = errors.New("failed to create profile using db")
	ErrGetProfileDB    = errors.New("failed to get profile using db")
	ErrUpdateProfileDB = errors.New("failed to update profile using db")

	ErrProfileNotFound = errors.New("profile not found")
)
