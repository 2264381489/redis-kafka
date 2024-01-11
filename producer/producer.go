package producer

import (
	"awesomeProject/redis"
	"context"
)

type Producer struct {
	Name   string
	client *redis.Client
	// 用户自定义的生产者配置参数
	opts *ProducerOptions
}

type ProducerOptions struct {
	// topic 可以缓存的消息长度，单位：条. 当消息条数超过此数值时，会把老消息踢出队列
	msgQueueLen int
}

type ProducerOption func(opts *ProducerOptions)

// 配置函数化
func WithMsgQueueLen(len int) ProducerOption {
	return func(opts *ProducerOptions) {
		opts.msgQueueLen = len
	}
}

// 默认为每个 topic 保留 500 条消息
func repairProducer(opts *ProducerOptions) {
	if opts.msgQueueLen <= 0 {
		opts.msgQueueLen = 500
	}
}

// 生产者 producer 的构造器函数
func NewProducer(client *redis.Client, opts ...ProducerOption) *Producer {
	p := Producer{
		client: client,
		opts:   &ProducerOptions{},
	}

	// 用函数为全局函数进行数据注入

	// 注入用户自定义的配置参数
	for _, opt := range opts {
		// 这个是调用
		opt(p.opts)
	}

	// 对非法的配置参数进行修复
	repairProducer(p.opts)

	// 返回 producer 实例
	return &p
}

// 生产一条消息
func (p *Producer) SendMsg(ctx context.Context, topic, key, val string) (string, error) {
	return p.client.XADD(ctx, topic, p.opts.msgQueueLen, key, val)
}
