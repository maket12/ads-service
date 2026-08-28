package middleware

import (
	"net/http"
	"strings"

	"github.com/maket12/ads-service/backend/authservice/api/proto/generated/auth_v1"
	"github.com/maket12/ads-service/backend/authservice/pkg/utils"
)

func WithAuth(authClient auth_v1.AuthServiceClient) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			authHandler := r.Header.Get("Authorization")
			token := strings.TrimPrefix(authHandler, "Bearer ")

			if token != "" {
				resp, err := authClient.ValidateAccessToken(ctx, &auth_v1.ValidateAccessTokenRequest{
					AccessToken: token,
				})
				if err == nil && resp != nil {
					ctx = utils.SetAccountIDInCtx(ctx, resp.AccountId)
					ctx = utils.SetAccountRoleInCtx(ctx, resp.Role)
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
