package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/maket12/ads-service/backend/adservice/api/proto/generated/ad_v1"
	"github.com/maket12/ads-service/backend/authservice/api/proto/generated/auth_v2"
	"github.com/maket12/ads-service/backend/gateway/cmd/app/config"
	"github.com/maket12/ads-service/backend/gateway/graph"
	"github.com/maket12/ads-service/backend/gateway/internal/middleware"
	"github.com/maket12/ads-service/backend/searchservice/api/proto/generated/search_v1"
	"github.com/maket12/ads-service/backend/userservice/api/proto/generated/user_v1"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

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

func closeSearchConnection(searchConn *grpc.ClientConn) {
	log.Printf("Gateway: Closing Search Service Connection...")
	if err := searchConn.Close(); err != nil {
		log.Printf(
			"Gateway: ERROR - could not close Search Service Connection: %v",
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
	authConn := mustDial(cfg.AuthGRPCAddr)
	defer closeAuthConnection(authConn)

	userConn := mustDial(cfg.UserGRPCAddr)
	defer closeUserConnection(userConn)

	adConn := mustDial(cfg.AdGRPCAddr)
	defer closeAdConnection(adConn)

	searchConn := mustDial(cfg.SearchGRPCAddr)
	defer closeSearchConnection(searchConn)

	authClient := auth_v2.NewAuthServiceClient(authConn)
	userClient := user_v1.NewUserServiceClient(userConn)
	adClient := ad_v1.NewAdServiceClient(adConn)
	searchClient := search_v1.NewSearchServiceClient(searchConn)

	// Create resolver
	resolver := &graph.Resolver{
		AuthClient:   authClient,
		UserClient:   userClient,
		AdClient:     adClient,
		SearchClient: searchClient,
	}

	// New GraphQL server
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{
		Resolvers: resolver,
	}))

	// Final handler with middleware
	mux := http.NewServeMux()
	mux.Handle("/", playground.Handler("GraphQL playground", "/query"))
	mux.Handle("/query", middleware.WithAuth(authClient)(srv))

	log.Printf("Gateway: Server is running on port %d", cfg.GatewayPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.GatewayPort), mux))
}

// mustDial creates a gRPC client connection with the auth-propagation
// interceptor attached, so every outgoing call forwards account_id/role
// from the Go context (set by middleware.Auth) into gRPC metadata.
func mustDial(addr string) *grpc.ClientConn {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(middleware.AuthPropagationInterceptor),
	)
	if err != nil {
		log.Fatalf("failed to dial %s: %v", addr, err)
	}
	return conn
}
