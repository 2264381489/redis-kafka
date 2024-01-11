package redis

import (
	conf "awesomeProject/config"
	"context"
	"github.com/google/uuid"
	"github.com/gookit/goutil/testutil/assert"
	"testing"
)

func TestClient_XReadGroup(t *testing.T) {
	ast := assert.New(t)
	ctx := context.Background()
	client := NewClient(ctx, &conf.RedisConfig{
		Addr:        "localhost:6379",
		Network:     "tcp",
		TimeOut:     1300,
		MaxIdle:     10,
		IdleTimeout: 1300,
		MaxActive:   20,
	})

	// XReadGroup 对于 > 会获取最新的数据
	groupID := "group_id:" + uuid.NewString()
	topicID := "topic_id:" + uuid.NewString()
	ast.Nil(client.XGROUP(ctx, groupID, topicID, nil))
	_, err := client.XADD(ctx, topicID, 1000, "test01", "test02")
	ast.Nil(err)
	_, err = client.XADD(ctx, topicID, 1000, "test01", "test02")
	ast.Nil(err)
	msgs, err := client.XReadGroup(ctx, groupID, "customer01", topicID, 0, false)
	ast.Nil(err)
	ast.Len(msgs, 2)
	for _, msg := range msgs {
		ast.Equal(msg.Key, "test01")
		ast.Equal(msg.Val, "test02")
	}
	ctx = context.Background()
	// 倘若 pending 为 false，代表需要消费的是尚未分配给任何 consumer 的新消息，此时会才用阻塞模式执行操作
	msgs, err = client.XReadGroup(ctx, groupID, "customer01", topicID, 13000, false)
	ast.NotNil(err) // 超时
	//  倘若 pending 为 true，代表需要消费的是已分配给当前 consumer 但是还未经 xack 确认的老消息. 此时采用非阻塞模式进行处理
	msgs, err = client.XReadGroup(ctx, groupID, "customer01", topicID, 13000, true)
	ast.Len(msgs, 2)
	for _, msg := range msgs {
		ast.Equal(msg.Key, "test01")
		ast.Equal(msg.Val, "test02")
	}

}
