package errs

import "errors"

/*
================ Validation failures ================
*/
var (
	ErrInvalidAdIndex = errors.New("invalid ad index id")
	ErrInvalidInput   = errors.New("invalid input") // for rich models
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
