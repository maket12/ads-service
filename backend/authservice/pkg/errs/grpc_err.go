package errs

import "google.golang.org/grpc/codes"

type OutErr struct {
	Code    codes.Code
	Message string
	Reason  error
}

func NewOutError(code codes.Code, msg string, reason error) *OutErr {
	return &OutErr{
		Code:    code,
		Message: msg,
		Reason:  reason,
	}
}
