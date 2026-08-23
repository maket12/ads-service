// Package middleware holds HTTP-layer middleware for the gateway.
package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	auth_v1 "github.com/maket12/ads-service/authservice/pkg/generated/auth_v1"

	"github.com/maket12/ads-service/gateway/internal/authctx"
)

// Auth resolves an "Authorization: Bearer <token>" header to an account
// id/role via AuthService.ValidateAccessToken and stashes it in the request
// context under authctx. Missing or invalid tokens are NOT rejected here —
// this gateway serves both public queries (e.g. an unauthenticated `ad`
// lookup) and authenticated ones off the same schema, so the decision to
// require auth belongs to each resolver via authctx.MustFromContext.
func Auth(authClient auth_v1.AuthServiceClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := bearerToken(r.Header.Get("Authorization"))
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			resp, err := authClient.ValidateAccessToken(r.Context(), &auth_v1.ValidateAccessTokenRequest{
				AccessToken: token,
			})
			if err != nil {
				// Invalid/expired token: treat the request as anonymous rather
				// than hard-failing the whole HTTP request, so a query that
				// doesn't need auth still succeeds. Resolvers that need a
				// caller will reject it via authctx.MustFromContext.
				slog.DebugContext(r.Context(), "access token validation failed", "err", err)
				next.ServeHTTP(w, r)
				return
			}

			ctx := authctx.WithIdentity(r.Context(), authctx.Identity{
				AccountID: resp.GetAccountId(),
				Role:      resp.GetRole(),
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if token == "" {
		return "", false
	}
	return token, true
}
