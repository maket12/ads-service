package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	AuthGRPCAddr string `env:"AUTH_GRPC_ADDR" envDefault:"localhost:50051"`
	UserGRPCAddr string `env:"USER_GRPC_ADDR" envDefault:"localhost:50052"`
	AdGRPCAddr   string `env:"AD_GRPC_ADDR" envDefault:"localhost:50053"`
	GatewayPort  int    `env:"GATEWAY_PORT" envDefault:"8080"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}
	return cfg, nil
}
