package consumer

import "time"

type ConsumerOptions struct {
	// 每轮阻塞消费新消息时等待超时时长
	receiveTimeout int
	// 处理消息的最大重试次数，超过此次数时，消息会被投递到死信队列
	maxRetryLimit int
	// 死信队列，可以由使用方自定义实现
	deadLetterMailbox DeadLetterMailbox
	// 投递死信流程超时阈值
	deadLetterDeliverTimeout time.Duration
	// 处理消息流程超时阈值
	handleMsgsTimeout time.Duration
}

type ConsumerOption func(opts *ConsumerOptions)

func WithReceiveTimeout(timeout int) ConsumerOption {
	return func(opts *ConsumerOptions) {
		opts.receiveTimeout = timeout
	}
}

func WithMaxRetryLimit(maxRetryLimit int) ConsumerOption {
	return func(opts *ConsumerOptions) {
		opts.maxRetryLimit = maxRetryLimit
	}
}

func WithDeadLetterMailbox(mailbox DeadLetterMailbox) ConsumerOption {
	return func(opts *ConsumerOptions) {
		opts.deadLetterMailbox = mailbox
	}
}

func WithDeadLetterDeliverTimeout(timeout time.Duration) ConsumerOption {
	return func(opts *ConsumerOptions) {
		opts.deadLetterDeliverTimeout = timeout
	}
}

func WithHandleMsgsTimeout(timeout time.Duration) ConsumerOption {
	return func(opts *ConsumerOptions) {
		opts.handleMsgsTimeout = timeout
	}
}

// 修复非法的配置参数
func repairConsumer(opts *ConsumerOptions) {
	// 默认阻塞消费的超时时长为 2 s
	if opts.receiveTimeout < 0 {
		opts.receiveTimeout = 2000
	}

	// 默认同一条消息最多处理 3 次，超过此次数，则进入死信
	if opts.maxRetryLimit < 0 {
		opts.maxRetryLimit = 3
	}

	// 如果用户没传入死信队列，则使用默认的死信队列，只会打印一下消息
	if opts.deadLetterMailbox == nil {
		opts.deadLetterMailbox = NewDeadLetterLogger()
	}

	// 投递消息进入死信队列的流程的超时时间
	if opts.deadLetterDeliverTimeout <= 0 {
		opts.deadLetterDeliverTimeout = time.Second
	}

	// 处理消息执行 callback 回调的超时时间
	if opts.handleMsgsTimeout <= 0 {
		opts.handleMsgsTimeout = time.Second
	}
}
