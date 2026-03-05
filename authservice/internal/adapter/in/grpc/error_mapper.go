package grpc

import (
	"errors"

	ucerrs "github.com/maket12/ads-service/authservice/internal/app/errs"
	pkgerrs "github.com/maket12/ads-service/pkg/errs"

	"google.golang.org/grpc/codes"
)

func gRPCError(err error) *pkgerrs.OutErr {
	var w *ucerrs.WrappedError
	if errors.As(err, &w) {
		switch {
		case errors.Is(w.Public, ucerrs.ErrHashPassword),
			errors.Is(w.Public, ucerrs.ErrCreateAccountDB),
			errors.Is(w.Public, ucerrs.ErrGetAccountByEmailDB),
			errors.Is(w.Public, ucerrs.ErrGetAccountByIDDB),
			errors.Is(w.Public, ucerrs.ErrUpdateAccountDB),
			errors.Is(w.Public, ucerrs.ErrGetAccountRoleDB),
			errors.Is(w.Public, ucerrs.ErrUpdateAccountRoleDB),
			errors.Is(w.Public, ucerrs.ErrCreateRefreshSessionDB),
			errors.Is(w.Public, ucerrs.ErrGetRefreshSessionByIDDB),
			errors.Is(w.Public, ucerrs.ErrRevokeRefreshSessionDB),
			errors.Is(w.Public, ucerrs.ErrCreateAccountRoleDB),
			errors.Is(w.Public, ucerrs.ErrGenerateAccessToken),
			errors.Is(w.Public, ucerrs.ErrGenerateRefreshToken),
			errors.Is(w.Public, ucerrs.ErrPublishEvent):
			return pkgerrs.NewOutError(codes.Internal, w.Public.Error(), w.Reason)

		case errors.Is(w.Public, ucerrs.ErrInvalidInput):
			return pkgerrs.NewOutError(codes.InvalidArgument, w.Public.Error(), w.Reason)

		default:
			return pkgerrs.NewOutError(codes.Internal, "internal error", w.Reason)
		}
	}

	switch {
	case errors.Is(err, ucerrs.ErrInvalidCredentials),
		errors.Is(err, ucerrs.ErrInvalidAccountID):
		return pkgerrs.NewOutError(codes.NotFound, err.Error(), nil)

	case errors.Is(err, ucerrs.ErrAccountAlreadyExists):
		return pkgerrs.NewOutError(codes.AlreadyExists, err.Error(), nil)

	case errors.Is(err, ucerrs.ErrCannotLogin),
		errors.Is(err, ucerrs.ErrCannotAssign),
		errors.Is(err, ucerrs.ErrCannotRevoke):
		return pkgerrs.NewOutError(codes.FailedPrecondition, err.Error(), nil)

	case errors.Is(err, ucerrs.ErrInvalidAccessToken),
		errors.Is(err, ucerrs.ErrInvalidRefreshToken):
		return pkgerrs.NewOutError(codes.Unauthenticated, err.Error(), nil)
	}

	return pkgerrs.NewOutError(codes.Internal, "internal error", nil)
}
