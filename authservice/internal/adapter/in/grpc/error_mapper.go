package grpc

import (
	"ads/authservice/internal/app/uc_errors"
	"ads/pkg/errs"
	"errors"

	"google.golang.org/grpc/codes"
)

func gRPCError(err error) *errs.OutErr {
	var w *uc_errors.WrappedError
	if errors.As(err, &w) {
		switch {
		case errors.Is(w.Public, uc_errors.ErrHashPassword),
			errors.Is(w.Public, uc_errors.ErrCreateAccountDB),
			errors.Is(w.Public, uc_errors.ErrGetAccountByEmailDB),
			errors.Is(w.Public, uc_errors.ErrGetAccountByIDDB),
			errors.Is(w.Public, uc_errors.ErrUpdateAccountDB),
			errors.Is(w.Public, uc_errors.ErrGetAccountRoleDB),
			errors.Is(w.Public, uc_errors.ErrUpdateAccountRoleDB),
			errors.Is(w.Public, uc_errors.ErrCreateRefreshSessionDB),
			errors.Is(w.Public, uc_errors.ErrGetRefreshSessionByIDDB),
			errors.Is(w.Public, uc_errors.ErrRevokeRefreshSessionDB),
			errors.Is(w.Public, uc_errors.ErrCreateAccountRoleDB),
			errors.Is(w.Public, uc_errors.ErrGenerateAccessToken),
			errors.Is(w.Public, uc_errors.ErrGenerateRefreshToken),
			errors.Is(w.Public, uc_errors.ErrPublishEvent):
			return errs.NewOutError(codes.Internal, w.Public.Error(), w.Reason)

		case errors.Is(w.Public, uc_errors.ErrInvalidInput):
			return errs.NewOutError(codes.InvalidArgument, w.Public.Error(), w.Reason)

		default:
			return errs.NewOutError(codes.Internal, "internal error", w.Reason)
		}
	}

	switch {
	case errors.Is(err, uc_errors.ErrInvalidCredentials),
		errors.Is(err, uc_errors.ErrInvalidAccountID):
		return errs.NewOutError(codes.NotFound, err.Error(), nil)

	case errors.Is(err, uc_errors.ErrAccountAlreadyExists):
		return errs.NewOutError(codes.AlreadyExists, err.Error(), nil)

	case errors.Is(err, uc_errors.ErrCannotLogin),
		errors.Is(err, uc_errors.ErrCannotAssign),
		errors.Is(err, uc_errors.ErrCannotRevoke):
		return errs.NewOutError(codes.FailedPrecondition, err.Error(), nil)

	case errors.Is(err, uc_errors.ErrInvalidAccessToken),
		errors.Is(err, uc_errors.ErrInvalidRefreshToken):
		return errs.NewOutError(codes.Unauthenticated, err.Error(), nil)
	}

	return errs.NewOutError(codes.Internal, "internal error", nil)
}
