package grpc

import (
	"errors"

	ucerrs "github.com/maket12/ads-service/adservice/internal/app/errs"
	pkgerrs "github.com/maket12/ads-service/adservice/pkg/errs"

	"google.golang.org/grpc/codes"
)

// Parses service error and returns response for grpc
func gRPCError(err error) *pkgerrs.OutErr {
	var w *ucerrs.WrappedError
	if errors.As(err, &w) {
		switch {
		case errors.Is(w.Public, ucerrs.ErrSaveImagesDB),
			errors.Is(w.Public, ucerrs.ErrGetImagesDB),
			errors.Is(w.Public, ucerrs.ErrDeleteImagesDB),
			errors.Is(w.Public, ucerrs.ErrCreateAdDB),
			errors.Is(w.Public, ucerrs.ErrGetAdDB),
			errors.Is(w.Public, ucerrs.ErrUpdateAdDB),
			errors.Is(w.Public, ucerrs.ErrDeleteAdDB),
			errors.Is(w.Public, ucerrs.ErrDeleteAllAdsDB),
			errors.Is(w.Public, ucerrs.ErrListAdsDB),
			errors.Is(w.Public, ucerrs.ErrListSellerAdsDB):
			return pkgerrs.NewOutError(codes.Internal, w.Public.Error(), w.Reason)

		case errors.Is(w.Public, ucerrs.ErrInvalidInput):
			return pkgerrs.NewOutError(codes.InvalidArgument, w.Public.Error(), w.Reason)

		default:
			return pkgerrs.NewOutError(codes.Internal, "internal error", w.Reason)
		}
	}

	switch {
	case errors.Is(err, ucerrs.ErrAccessDenied):
		return pkgerrs.NewOutError(codes.PermissionDenied, err.Error(), nil)

	case errors.Is(err, ucerrs.ErrAdNotFound):
		return pkgerrs.NewOutError(codes.NotFound, err.Error(), nil)

	case errors.Is(err, ucerrs.ErrCannotPublish),
		errors.Is(err, ucerrs.ErrCannotReject),
		errors.Is(err, ucerrs.ErrCannotDelete):
		return pkgerrs.NewOutError(codes.FailedPrecondition, err.Error(), nil)

	case errors.Is(err, pkgerrs.ErrNotAuthenticated):
		return pkgerrs.NewOutError(codes.Unauthenticated, err.Error(), nil)
	}

	return pkgerrs.NewOutError(codes.Internal, "internal error", nil)
}
