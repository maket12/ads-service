//go:build integration

package redis_test

import (
	"context"

	pkgredis "github.com/maket12/ads-service/authservice/pkg/redis"
	"github.com/stretchr/testify/suite"
)

type BaseRepoSuite struct {
	suite.Suite
	ctx            context.Context
	redisContainer *pkgredis.TestContainer
	redisClient    *pkgredis.Client
}

func (s *BaseRepoSuite) SetupBase() {
	s.ctx = context.Background()

	var err error
	s.redisContainer, err = pkgredis.StartTestContainer(s.ctx, nil)
	s.Require().NoError(err)

	s.redisClient, err = pkgredis.NewClient(s.ctx, s.redisContainer.Config)
	s.Require().NoError(err)
}

func (s *BaseRepoSuite) TearDownSuite() {
	if s.redisClient != nil {
		_ = s.redisClient.Close()
	}
	if s.redisContainer != nil {
		_ = s.redisContainer.Close(s.ctx)
	}
}
