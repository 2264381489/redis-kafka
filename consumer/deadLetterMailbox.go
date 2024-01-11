package consumer

import (
	"awesomeProject/redis"
	"context"
	"log"
)

// 死信队列，当消息处理失败达到指定次数时，会被投递到此处
type DeadLetterMailbox interface {
	Deliver(ctx context.Context, msg *redis.MsgEntity) error
}

// 默认使用的死信队列，仅仅对消息失败的信息进行日志打印
type DeadLetterLogger struct {
	log *log.Logger
}

func NewDeadLetterLogger() *DeadLetterLogger {
	return &DeadLetterLogger{
		log: log.Default(),
	}
}

func (d *DeadLetterLogger) Deliver(ctx context.Context, msg *redis.MsgEntity) error {
	d.log.Printf("msg fail execeed retry limit, msg id: %s", msg.MsgID)
	return nil
}
