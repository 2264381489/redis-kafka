package logic

import (
	"awesomeProject/svc"
	"context"
	"log"
)

type PingLogic struct {
	log    *log.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PingLogic {
	return &PingLogic{
		log:    log.Default(),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (c *PingLogic) Ping() (string, error) {
	conn, err := c.svcCtx.RedisClient.GetConn(c.ctx)
	if err != nil {
		return "", err
	}
	reply, err := conn.Do("ping")
	if err != nil {
		return "", err
	}
	reply, err = c.svcCtx.RedisClient.XADD(c.ctx, "test", 0, "test", "test2")
	return reply.(string), nil
}
