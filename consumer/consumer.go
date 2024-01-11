package consumer

import (
	"awesomeProject/redis"
	"context"
	"log"
)

// 消费者 consumer 类定义
type Consumer struct {
	// consumer 生命周期管理
	ctx context.Context
	// 停止 consumer 的控制器
	stop context.CancelFunc

	// 接收到 msg 时执行的回调函数，由使用方定义
	callbackFunc MsgCallback

	// redis 客户端，基于 redis 实现 message queue
	client *redis.Client

	// 消费的 topic
	topic string
	// 所属的消费者组
	groupID string
	// 当前节点的消费者 id
	consumerID string

	// 各消息累计失败次数
	failureCnts map[redis.MsgEntity]int

	// 一些用户自定义的配置
	opts *ConsumerOptions

	DeadLetterMailbox DeadLetterMailbox
}

type MsgCallback func(ctx context.Context, msg *redis.MsgEntity) error

// 消费者 consumer 构造器函数
func NewConsumer(client *redis.Client, topic, groupID, consumerID string, callbackFunc MsgCallback, opts ...ConsumerOption) (*Consumer, error) {

	// cancel context，用于提供停止 consumer 的控制器
	ctx, stop := context.WithCancel(context.Background())
	// 构造 consumer 实例
	c := Consumer{
		client:       client,
		ctx:          ctx,
		stop:         stop,
		callbackFunc: callbackFunc,
		topic:        topic,
		groupID:      groupID,
		consumerID:   consumerID,

		opts: &ConsumerOptions{},

		failureCnts: make(map[redis.MsgEntity]int),
	}

	// 校验 consumer 中的参数，包括 topic、groupID、consumerID 都不能为空，且 callbackFunc 不能为空
	if err := c.checkParam(); err != nil {
		return nil, err
	}

	// 注入用户自定义的配置项
	for _, opt := range opts {
		opt(c.opts)
	}

	// 修复非法的配置参数
	repairConsumer(c.opts)

	// 启动 consumer 守护 goroutine，负责轮询消费消息
	go c.run()
	// 返回 consumer 实例
	return &c, nil
}

func (c *Consumer) run() {
	// 自旋模型
	for {
		select {
		// 退出自旋
		case <-c.ctx.Done():
			return
		default:
		}

		// 获取消息
		msgs, err := c.receiveMsg()

		if err != nil {
			continue
		}
		c.handlerMsg(msgs)

		c.deliverDeadMailBox()

	}
}

func (c *Consumer) deliverDeadMailBox() {
	tctx, _ := context.WithTimeout(c.ctx, c.opts.handleMsgsTimeout)
	for entity, cnt := range c.failureCnts {
		if cnt < c.opts.maxRetryLimit {
			continue
		}

		// 超过最大重试次数，进入死信队列
		if err := c.DeadLetterMailbox.Deliver(tctx, &entity); err != nil {
			log.Print(c.ctx, "dead letter deliver failed, msg id: %s, err: %v", entity.MsgID, err)
			continue
		}

		delete(c.failureCnts, entity)
	}
}

func (c *Consumer) handlerMsg(msgs []*redis.MsgEntity) {
	for _, msg := range msgs {
		tctx, _ := context.WithTimeout(c.ctx, c.opts.handleMsgsTimeout)
		// 处理消息
		if err := c.callbackFunc(tctx, msg); err != nil {
			// 执行失败计入失败次数
			c.failureCnts[*msg]++
			continue
		}
		// 执行成功 ack
		_, err := c.client.XACK(c.ctx, c.topic, c.groupID, msg.MsgID)
		if err != nil {
			continue
		}
		// 无须考虑发送成功的情况，依然删除
		delete(c.failureCnts, *msg)
	}
}

func (c *Consumer) receiveMsg() ([]*redis.MsgEntity, error) {
	// receiveMsg 方法调用 redis 的 XReadGroup 方法 同步获取最新的
	return c.client.XReadGroup(c.ctx, c.groupID, c.consumerID, c.topic, c.opts.receiveTimeout, false)

}

func (c *Consumer) checkParam() error {

	return nil
}
