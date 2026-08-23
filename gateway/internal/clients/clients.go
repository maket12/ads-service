// Package clients dials the three downstream gRPC services and exposes
// their typed clients as a single bundle that gets threaded through the
// resolver via graph.Resolver.
package clients

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	ad_v1 "github.com/maket12/ads-service/adservice/pkg/generated/ad_v1"
	auth_v1 "github.com/maket12/ads-service/authservice/pkg/generated/auth_v1"
	user_v1 "github.com/maket12/ads-service/userservice/pkg/generated/user_v1"
)

// Config holds the dial targets for each downstream service, e.g.
// "authservice:9091" in-cluster or "localhost:9091" locally.
type Config struct {
	AuthAddr string
	UserAddr string
	AdAddr   string
}

// Clients bundles the typed gRPC clients used by resolvers.
type Clients struct {
	Auth auth_v1.AuthServiceClient
	User user_v1.UserServiceClient
	Ad   ad_v1.AdServiceClient

	conns []*grpc.ClientConn
}

// Dial connects to all three downstream services. Connections are lazy
// (grpc.NewClient does not block until first RPC), so this returns quickly
// and failures surface as normal RPC errors from resolvers, not at startup.
func Dial(cfg Config) (*Clients, error) {
	authConn, err := dial(cfg.AuthAddr)
	if err != nil {
		return nil, fmt.Errorf("dial authservice: %w", err)
	}
	userConn, err := dial(cfg.UserAddr)
	if err != nil {
		return nil, fmt.Errorf("dial userservice: %w", err)
	}
	adConn, err := dial(cfg.AdAddr)
	if err != nil {
		return nil, fmt.Errorf("dial adservice: %w", err)
	}

	return &Clients{
		Auth:  auth_v1.NewAuthServiceClient(authConn),
		User:  user_v1.NewUserServiceClient(userConn),
		Ad:    ad_v1.NewAdServiceClient(adConn),
		conns: []*grpc.ClientConn{authConn, userConn, adConn},
	}, nil
}

func dial(addr string) (*grpc.ClientConn, error) {
	// TODO: swap insecure.NewCredentials() for TLS creds before shipping
	// this past a trusted internal network.
	return grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

// Close tears down all downstream connections. Call on server shutdown.
func (c *Clients) Close() error {
	var firstErr error
	for _, conn := range c.conns {
		if err := conn.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
