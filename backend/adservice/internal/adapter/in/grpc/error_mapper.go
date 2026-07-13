package grpc

import (
	"errors"

	"github.com/maket12/ads-service/backend/adservice/internal/app/errs"
	pkgerrs "github.com/maket12/ads-service/pkg/errs"

	"google.golang.org/grpc/codes"
)

// Parses service error and returns response for grpc
func gRPCError(err error) *pkgerrs.OutErr {
	var w *errs.WrappedError
	if errors.As(err, &w) {
		switch {
		case errors.Is(w.Public, errs.ErrSaveImagesDB),
			errors.Is(w.Public, errs.ErrGetImagesDB),
			errors.Is(w.Public, errs.ErrDeleteImagesDB),
			errors.Is(w.Public, errs.ErrCreateAdDB),
			errors.Is(w.Public, errs.ErrGetAdDB),
			errors.Is(w.Public, errs.ErrUpdateAdDB),
			errors.Is(w.Public, errs.ErrUpdateAdStatusDB),
			errors.Is(w.Public, errs.ErrDeleteAdDB),
			errors.Is(w.Public, errs.ErrDeleteAllAdsDB):
			return pkgerrs.NewOutError(codes.Internal, w.Public.Error(), w.Reason)

		case errors.Is(w.Public, errs.ErrInvalidInput):
			return pkgerrs.NewOutError(codes.InvalidArgument, w.Public.Error(), w.Reason)

		default:
			return pkgerrs.NewOutError(codes.Internal, "internal error", w.Reason)
		}
	}

	switch {
	case errors.Is(err, errs.ErrAccessDenied):
		return pkgerrs.NewOutError(codes.PermissionDenied, err.Error(), nil)

	case errors.Is(err, errs.ErrInvalidAdID):
		return pkgerrs.NewOutError(codes.NotFound, err.Error(), nil)

	case errors.Is(err, errs.ErrCannotPublish),
		errors.Is(err, errs.ErrCannotReject),
		errors.Is(err, errs.ErrCannotDelete):
		return pkgerrs.NewOutError(codes.FailedPrecondition, err.Error(), nil)
	}

	return pkgerrs.NewOutError(codes.Internal, "internal error", nil)
}
