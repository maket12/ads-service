package errs

import (
	"log/slog"

	"google.golang.org/grpc/codes"
)

type OutErr struct {
	Code    codes.Code
	Message string
	Reason  error
	Level   slog.Level
}

func NewOutError(code codes.Code, msg string, reason error) *OutErr {
	return &OutErr{
		Code:    code,
		Message: msg,
		Reason:  reason,
		Level:   levelForCode(code),
	}
}

func levelForCode(code codes.Code) slog.Level {
	switch code {
	case codes.InvalidArgument, codes.NotFound, codes.AlreadyExists,
		codes.Unauthenticated, codes.PermissionDenied:
		return slog.LevelWarn
	default:
		return slog.LevelError
	}
}
