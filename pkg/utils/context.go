package utils

import (
	"context"
	"errors"
	pkgerrs "github.com/maket12/ads-service/pkg/errs"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

// Context keys
type contextKey string

const (
	AccountIDKey   contextKey = "account_id"
	AccountRoleKey contextKey = "account_role"
)

// Custom errors
var (
	ErrMetadataIsMissing       = errors.New("metadata is missing")
	ErrAccountIDNotSpecified   = errors.New("account id not found in metadata")
	ErrInvalidAccountID        = errors.New("metadata contains invalid account id")
	ErrAccountRoleNotSpecified = errors.New("account role not found in metadata")
)

// ExtractAccountID Extracts account id from incoming context (GRPC)
func ExtractAccountID(ctx context.Context) (uuid.UUID, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return uuid.Nil, pkgerrs.NewNotAuthenticatedErrorWithReason(ErrMetadataIsMissing)
	}

	vals := md.Get("x-account-id")
	if len(vals) == 0 {
		return uuid.Nil, pkgerrs.NewNotAuthenticatedErrorWithReason(ErrAccountIDNotSpecified)
	}

	id, err := uuid.Parse(vals[0])
	if err != nil {
		return uuid.Nil, pkgerrs.NewNotAuthenticatedErrorWithReason(ErrInvalidAccountID)
	}

	return id, nil
}

// ExtractAccountRole Extracts account role from incoming context (GRPC)
func ExtractAccountRole(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", pkgerrs.NewNotAuthenticatedErrorWithReason(ErrMetadataIsMissing)
	}

	vals := md.Get("x-account-role")
	if len(vals) == 0 {
		return "", pkgerrs.NewNotAuthenticatedErrorWithReason(ErrAccountRoleNotSpecified)
	}

	return vals[0], nil
}

// PackAccountIDForGRPC Packs account id into outgoing context (metadata | GRPC)
func PackAccountIDForGRPC(ctx context.Context, accountID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "x-account-id", accountID)
}

// PackAccountRoleForGRPC Packs account role into outgoing context (metadata | GRPC)
func PackAccountRoleForGRPC(ctx context.Context, accountRole string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "x-account-role", accountRole)
}

// SetAccountIDInCtx Sets account id in context (gateway)
func SetAccountIDInCtx(ctx context.Context, accountID string) context.Context {
	return context.WithValue(ctx, AccountIDKey, accountID)
}

// SetAccountRoleInCtx Sets account role in context (gateway)
func SetAccountRoleInCtx(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, AccountRoleKey, role)
}
