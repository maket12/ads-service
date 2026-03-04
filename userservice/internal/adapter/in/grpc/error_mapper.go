package grpc

import (
	"ads/pkg/errs"
	"ads/userservice/internal/app/uc_errors"
	"errors"

	"google.golang.org/grpc/codes"
)

// Parses service error and returns response for grpc
func gRPCError(err error) *errs.OutErr {
	var w *uc_errors.WrappedError
	if errors.As(err, &w) {
		switch {
		case errors.Is(w.Public, uc_errors.ErrCreateProfileDB),
			errors.Is(w.Public, uc_errors.ErrGetProfileDB),
			errors.Is(w.Public, uc_errors.ErrUpdateProfileDB):
			return errs.NewOutError(codes.Internal, w.Public.Error(), w.Reason)

		default:
			return errs.NewOutError(codes.Internal, "internal error", w.Reason)
		}
	}

	switch {
	case errors.Is(err, uc_errors.ErrInvalidAccountID):
		return errs.NewOutError(codes.NotFound, err.Error(), nil)

	case errors.Is(err, uc_errors.ErrInvalidProfileData),
		errors.Is(err, uc_errors.ErrInvalidPhoneNumber):
		return errs.NewOutError(codes.InvalidArgument, err.Error(), nil)

	case errors.Is(err, errs.ErrNotAuthenticated):
		return errs.NewOutError(codes.Unauthenticated, err.Error(), nil)
	}

	return errs.NewOutError(codes.Internal, "internal error", nil)
}
