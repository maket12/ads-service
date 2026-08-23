package authctx

import "errors"

// ErrUnauthenticated is returned by resolvers when a mutation/query that
// requires a logged-in caller is hit without a valid access token.
var ErrUnauthenticated = errors.New("authentication required")

// ErrForbidden is returned by resolvers that enforce a role requirement
// (e.g. assignRole) when the caller's role doesn't qualify.
var ErrForbidden = errors.New("forbidden")
