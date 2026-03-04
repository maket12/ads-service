package main

import (
	"ads/gateway/cmd/app/config"
	"ads/gateway/graph"
	"ads/pkg/generated/ad_v1"
	"ads/pkg/generated/auth_v1"
	"ads/pkg/generated/user_v1"
	"ads/pkg/utils"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func AuthMiddleware(authClient auth_v1.AuthServiceClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				next.ServeHTTP(w, r)
				return
			}

			token := parts[1]

			resp, err := authClient.ValidateAccessToken(r.Context(), &auth_v1.ValidateAccessTokenRequest{
				AccessToken: token,
			})

			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := utils.SetAccountIDInCtx(r.Context(), resp.GetAccountId())
			ctx = utils.SetAccountRoleInCtx(ctx, resp.GetRole())

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func closeAuthConnection(authConn *grpc.ClientConn) {
	log.Printf("Gateway: Closing Auth Service Connection...")
	if err := authConn.Close(); err != nil {
		log.Printf(
			"Gateway: ERROR - could not close Auth Service Connection: %v",
			err,
		)
	}
}

func closeUserConnection(userConn *grpc.ClientConn) {
	log.Printf("Gateway: Closing User Service Connection...")
	if err := userConn.Close(); err != nil {
		log.Printf(
			"Gateway: ERROR - could not close User Service Connection: %v",
			err,
		)
	}
}

func closeAdConnection(adConn *grpc.ClientConn) {
	log.Printf("Gateway: Closing Ad Service Connection...")
	if err := adConn.Close(); err != nil {
		log.Printf(
			"Gateway: ERROR - could not close Ad Service Connection: %v",
			err,
		)
	}
}

func main() {
	// Load Config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Gateway: %v", err)
	}

	// Make connections to services
	authConn, err := grpc.NewClient(
		cfg.AuthGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("Gateway: WARNING - could not connect to Auth Service: %v", err)
	}
	defer closeAuthConnection(authConn)

	userConn, err := grpc.NewClient(
		cfg.UserGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("Gateway: WARNING - could not connect to User Service: %v", err)
	}
	defer closeUserConnection(userConn)

	addConn, err := grpc.NewClient(
		cfg.AdGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("Gateway: WARNING - could not connect to Ad Service: %v", err)
	}
	defer closeAdConnection(addConn)

	// Create resolver
	resolver := &graph.Resolver{
		AuthClient: auth_v1.NewAuthServiceClient(authConn),
		UserClient: user_v1.NewUserServiceClient(userConn),
		AdClient:   ad_v1.NewAdServiceClient(addConn),
	}

	// New GraphQL server
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{
		Resolvers: resolver,
	}))

	// Final handler with middleware
	router := AuthMiddleware(resolver.AuthClient)(srv)

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", router)

	log.Printf("Gateway: Server is running on port %d", cfg.GatewayPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.GatewayPort), nil))
}
