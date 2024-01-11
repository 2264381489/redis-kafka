package producer

import (
	conf "awesomeProject/config"
	"awesomeProject/redis"
	"context"
	"github.com/google/uuid"
	"github.com/gookit/goutil/testutil/assert"
	"testing"
)

func TestProducer_SendMsg(t *testing.T) {

	ast := assert.New(t)
	ctx := context.Background()
	client := redis.NewClient(ctx, &conf.RedisConfig{
		Addr:        "localhost:6379",
		Network:     "tcp",
		TimeOut:     1300,
		MaxIdle:     10,
		IdleTimeout: 1300,
		MaxActive:   20,
	})

	producer := NewProducer(client)
	topicID := "topic_id:" + uuid.NewString()
	msg, err := producer.SendMsg(ctx, topicID, "test", "val")
	ast.Nil(err)
	ast.NotEmpty(msg)
}
