package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/maket12/ads-service/gateway/graph"
	"github.com/maket12/ads-service/gateway/graph/generated"
	"github.com/maket12/ads-service/gateway/internal/clients"
	"github.com/maket12/ads-service/gateway/internal/middleware"
)

func main() {
	cfg := clients.Config{
		AuthAddr: envOr("AUTH_SERVICE_ADDR", "localhost:9091"),
		UserAddr: envOr("USER_SERVICE_ADDR", "localhost:9092"),
		AdAddr:   envOr("AD_SERVICE_ADDR", "localhost:9093"),
	}

	c, err := clients.Dial(cfg)
	if err != nil {
		slog.Error("failed to connect to downstream services", "err", err)
		os.Exit(1)
	}
	defer c.Close()

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: &graph.Resolver{Clients: c},
	}))
	srv.Use(extension.Introspection{})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	mux := http.NewServeMux()
	mux.Handle("/", playground.Handler("GraphQL playground", "/query"))
	mux.Handle("/query", middleware.Auth(c.Auth)(srv))

	addr := envOr("GATEWAY_ADDR", ":8080")
	slog.Info("gateway listening", "addr", addr,
		"authservice", cfg.AuthAddr, "userservice", cfg.UserAddr, "adservice", cfg.AdAddr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
