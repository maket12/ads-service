// Package authctx carries the identity of the calling user through a
// request's context.Context, from the HTTP auth middleware down into
// resolvers and out again as outgoing gRPC metadata to the downstream
// services.
//
// None of the downstream RPCs that need to know "who is calling" take an
// account_id field on the request (GetProfileRequest, CreateAdRequest, ...).
// That's intentional: the gateway is the only component that talks to
// AuthService, so it's the gateway's job to resolve the bearer token to an
// account id/role once and forward it as metadata on every subsequent call.
package authctx

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// Metadata header names used on outgoing gRPC calls to downstream services.
const (
	MetadataAccountID = "x-account-id"
	MetadataRole      = "x-role"
)

type identityKey struct{}

// Identity is the caller's identity as resolved from their access token.
type Identity struct {
	AccountID string
	Role      string
}

// WithIdentity returns a new context carrying the given identity.
func WithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, identityKey{}, id)
}

// FromContext returns the caller's identity, if the request was
// authenticated. ok is false for anonymous requests.
func FromContext(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(identityKey{}).(Identity)
	return id, ok
}

// MustFromContext returns the caller's identity or an error, for resolvers
// that require authentication.
func MustFromContext(ctx context.Context) (Identity, error) {
	id, ok := FromContext(ctx)
	if !ok {
		return Identity{}, ErrUnauthenticated
	}
	return id, nil
}

// Outgoing attaches the caller's identity (if any) as gRPC metadata on ctx,
// so downstream services can read x-account-id / x-role. Call this right
// before every downstream gRPC call made from a resolver.
func Outgoing(ctx context.Context) context.Context {
	id, ok := FromContext(ctx)
	if !ok {
		return ctx
	}
	md := metadata.Pairs(
		MetadataAccountID, id.AccountID,
		MetadataRole, id.Role,
	)
	return metadata.NewOutgoingContext(ctx, md)
}
