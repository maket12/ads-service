package errs

import "errors"

var (
	ErrAccessDenied = errors.New("no permission to access this data")
)

/*
================ Validation failures ================
*/
var (
	ErrInvalidInput  = errors.New("invalid input")
	ErrCannotPublish = errors.New("ad has been already published or not available")
	ErrCannotReject  = errors.New("ad has been already published or not available")
	ErrCannotDelete  = errors.New("ad has been already deleted or rejected")
)

/*
================ MongoDB failures ================
*/
var (
	ErrSaveImagesDB   = errors.New("failed to save images using db")
	ErrGetImagesDB    = errors.New("failed to get images using db")
	ErrDeleteImagesDB = errors.New("failed to delete images using db")
)

/*
================ Postgres failures ================
*/
var (
	ErrCreateAdDB      = errors.New("failed to create ad using db")
	ErrGetAdDB         = errors.New("failed to get add using db")
	ErrUpdateAdDB      = errors.New("failed to update ad using db")
	ErrDeleteAdDB      = errors.New("failed to delete ad using db")
	ErrDeleteAllAdsDB  = errors.New("failed to all ads using db")
	ErrListAdsDB       = errors.New("failed to get a list of ads using db")
	ErrListSellerAdsDB = errors.New("failed to get a list of ads by seller id using db")

	ErrAdNotFound = errors.New("ad not found")
)
