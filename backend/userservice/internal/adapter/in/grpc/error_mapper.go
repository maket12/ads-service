package grpc

import (
	"errors"

	ucerrs "github.com/maket12/ads-service/userservice/internal/app/errs"
	pkgerrs "github.com/maket12/ads-service/userservice/pkg/errs"

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

		case errors.Is(w.Public, ucerrs.ErrInvalidInput):
			return pkgerrs.NewOutError(
				codes.InvalidArgument,
				w.Public.Error(),
				w.Reason,
			)

		default:
			return pkgerrs.NewOutError(codes.Internal, "internal error", w.Reason)
		}
	}

	switch {
	case errors.Is(err, ucerrs.ErrProfileNotFound):
		return pkgerrs.NewOutError(codes.NotFound, err.Error(), nil)

	case errors.Is(err, pkgerrs.ErrNotAuthenticated):
		return pkgerrs.NewOutError(codes.Unauthenticated, err.Error(), nil)
	}

	return pkgerrs.NewOutError(codes.Internal, "internal error", nil)
}
