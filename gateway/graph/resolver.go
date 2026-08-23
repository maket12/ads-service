package graph

import "github.com/maket12/ads-service/gateway/internal/clients"

//go:generate go run github.com/99designs/gqlgen generate

// Resolver is the root resolver. It holds nothing but a handle to the
// downstream gRPC clients — the gateway keeps no state of its own.
type Resolver struct {
	Clients *clients.Clients
}
