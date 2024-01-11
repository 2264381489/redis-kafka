package svc

import (
	"awesomeProject/config"
	"awesomeProject/redis"
	"context"
)

type ServiceContext struct {
	Ctx         context.Context
	Config      config.Config
	RedisClient *redis.Client
}

func NewServiceContext(ctx context.Context, c config.Config) *ServiceContext {
	return &ServiceContext{
		Ctx:         ctx,
		Config:      c,
		RedisClient: redis.NewClient(ctx, &c.RedisConfig),
	}
}
