package grpc

import (
	"ads/adservice/internal/app/uc_errors"
	"ads/pkg/errs"
	"errors"

	"google.golang.org/grpc/codes"
)

// Parses service error and returns response for grpc
func gRPCError(err error) *errs.OutErr {
	var w *uc_errors.WrappedError
	if errors.As(err, &w) {
		switch {
		case errors.Is(w.Public, uc_errors.ErrSaveImagesDB),
			errors.Is(w.Public, uc_errors.ErrGetImagesDB),
			errors.Is(w.Public, uc_errors.ErrDeleteImagesDB),
			errors.Is(w.Public, uc_errors.ErrCreateAdDB),
			errors.Is(w.Public, uc_errors.ErrGetAdDB),
			errors.Is(w.Public, uc_errors.ErrUpdateAdDB),
			errors.Is(w.Public, uc_errors.ErrUpdateAdStatusDB),
			errors.Is(w.Public, uc_errors.ErrDeleteAdDB),
			errors.Is(w.Public, uc_errors.ErrDeleteAllAdsDB):
			return errs.NewOutError(codes.Internal, w.Public.Error(), w.Reason)

		case errors.Is(w.Public, uc_errors.ErrInvalidInput):
			return errs.NewOutError(codes.InvalidArgument, w.Public.Error(), w.Reason)

		default:
			return errs.NewOutError(codes.Internal, "internal error", w.Reason)
		}
	}

	switch {
	case errors.Is(err, uc_errors.ErrAccessDenied):
		return errs.NewOutError(codes.PermissionDenied, err.Error(), nil)

	case errors.Is(err, uc_errors.ErrInvalidAdID):
		return errs.NewOutError(codes.NotFound, err.Error(), nil)

	case errors.Is(err, uc_errors.ErrCannotPublish),
		errors.Is(err, uc_errors.ErrCannotReject),
		errors.Is(err, uc_errors.ErrCannotDelete):
		return errs.NewOutError(codes.FailedPrecondition, err.Error(), nil)
	}

	return errs.NewOutError(codes.Internal, "internal error", nil)
}
