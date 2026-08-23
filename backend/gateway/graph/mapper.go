package graph

import (
	"github.com/vektah/gqlparser/v2/gqlerror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapGRPCError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return err
	}
	switch st.Code() {
	case codes.InvalidArgument:
		return &gqlerror.Error{Message: st.Message(), Extensions: map[string]interface{}{"code": "INVALID_ARGUMENT"}}
	case codes.NotFound:
		return &gqlerror.Error{Message: st.Message(), Extensions: map[string]interface{}{"code": "NOT_FOUND"}}
	case codes.AlreadyExists:
		return &gqlerror.Error{Message: st.Message(), Extensions: map[string]interface{}{"code": "ALREADY_EXISTS"}}
	case codes.PermissionDenied:
		return &gqlerror.Error{Message: st.Message(), Extensions: map[string]interface{}{"code": "FORBIDDEN"}}
	case codes.FailedPrecondition:
		return &gqlerror.Error{Message: st.Message(), Extensions: map[string]interface{}{"code": "FAILED_PRECONDITION"}}
	case codes.Unauthenticated:
		return &gqlerror.Error{Message: st.Message(), Extensions: map[string]interface{}{"code": "UNAUTHENTICATED"}}
	default:
		return &gqlerror.Error{Message: "internal error", Extensions: map[string]interface{}{"code": "INTERNAL"}}
	}
}
