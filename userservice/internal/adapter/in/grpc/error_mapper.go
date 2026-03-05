package grpc

import (
	"errors"

	pkgerrs "github.com/maket12/ads-service/pkg/errs"
	ucerrs "github.com/maket12/ads-service/userservice/internal/app/errs"

	"google.golang.org/grpc/codes"
)

// Parses service error and returns response for grpc
func gRPCError(err error) *pkgerrs.OutErr {
	var w *ucerrs.WrappedError
	if errors.As(err, &w) {
		switch {
		case errors.Is(w.Public, ucerrs.ErrCreateProfileDB),
			errors.Is(w.Public, ucerrs.ErrGetProfileDB),
			errors.Is(w.Public, ucerrs.ErrUpdateProfileDB):
			return pkgerrs.NewOutError(codes.Internal, w.Public.Error(), w.Reason)

		default:
			return pkgerrs.NewOutError(codes.Internal, "internal error", w.Reason)
		}
	}

	switch {
	case errors.Is(err, ucerrs.ErrInvalidAccountID):
		return pkgerrs.NewOutError(codes.NotFound, err.Error(), nil)

	case errors.Is(err, ucerrs.ErrInvalidProfileData),
		errors.Is(err, ucerrs.ErrInvalidPhoneNumber):
		return pkgerrs.NewOutError(codes.InvalidArgument, err.Error(), nil)

	case errors.Is(err, pkgerrs.ErrNotAuthenticated):
		return pkgerrs.NewOutError(codes.Unauthenticated, err.Error(), nil)
	}

	return pkgerrs.NewOutError(codes.Internal, "internal error", nil)
}
