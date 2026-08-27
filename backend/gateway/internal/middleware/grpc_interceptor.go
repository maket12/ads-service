package middleware

import (
	"context"

	"github.com/maket12/ads-service/backend/authservice/pkg/utils"
	"google.golang.org/grpc"
)

func AuthPropagationInterceptor(
	ctx context.Context,
	method string,
	req, reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	if accountID, ok := ctx.Value(utils.AccountIDKey).(string); ok && accountID != "" {
		ctx = utils.PackAccountIDForGRPC(ctx, accountID)
	}
	if role, ok := ctx.Value(utils.AccountRoleKey).(string); ok && role != "" {
		ctx = utils.PackAccountRoleForGRPC(ctx, role)
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}
