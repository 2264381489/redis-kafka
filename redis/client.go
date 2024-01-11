package redis

import (
	conf "awesomeProject/config"
	"context"
	"errors"
	"github.com/demdxx/gocast"
	"github.com/gomodule/redigo/redis"
	"log"
	"time"
)

func dial(ctx context.Context, network, addr string, timeout int64) (redis.Conn, error) {
	return redis.DialContext(ctx, network, addr, redis.DialReadTimeout(time.Duration(timeout)*time.Millisecond), redis.DialClientName("test"))
}

type Client struct {
	ctx context.Context
	// 用户自定义配置
	opts *conf.RedisConfig
	// redis 连接池
	pool *redis.Pool
}

func NewClient(ctx context.Context, opts *conf.RedisConfig) *Client {

	c := &Client{
		ctx:  ctx,
		opts: opts,
	}
	pool := c.getRedisPool()
	c.pool = pool
	return c
}

func (c *Client) getRedisPool() *redis.Pool {
	return &redis.Pool{
		// 最大空闲连接数
		MaxIdle: c.opts.MaxIdle,
		// 连接最长空闲时间
		IdleTimeout: time.Duration(c.opts.IdleTimeout),
		// Dial 方法
		Dial: func() (redis.Conn, error) {
			conn, err := dial(c.ctx, c.opts.Network, c.opts.Addr, c.opts.TimeOut)
			if err != nil {
				log.Fatalf("redis connect err :%+v", err)
				return nil, err
			}
			// ping 没必要校验是否有pong
			_, err = conn.Do("ping")
			if err != nil {
				log.Fatalf("redis ping err :%+v", err)
				return nil, err
			}
			return conn, nil
		},
		MaxActive: c.opts.MaxActive,
		// 当连接不够时是阻塞 还是 立即返回错误
		Wait: c.opts.Wait,
		// 测试方法
		TestOnBorrow: func(conn redis.Conn, t time.Time) error {
			_, err := conn.Do("ping")
			if err != nil {
				log.Fatalf("redis ping err :%+v", err)
				return err
			}
			return nil
		},
	}
}

// 获取连接
func (c *Client) GetConn(ctx context.Context) (redis.Conn, error) {
	return c.pool.GetContext(ctx)
}

// 投递消息
func (c *Client) XADD(ctx context.Context, topic string, maxLen int, value ...string) (string, error) {

	if topic == "" {
		return "", errors.New("topic is invalid")
	}
	// 从 redis 连接池中获取连接
	conn, err := c.pool.GetContext(ctx)
	if err != nil {
		return "", err
	}
	// 使用完毕后把连接放回连接池
	defer conn.Close()
	var command []interface{}
	command = append(command, topic, "MAXLEN", maxLen, "*")

	for _, val := range value {
		command = append(command, val)
	}

	//return redis.String(conn.Do("XADD", topic, "*", "123", "456", "789")) 似乎不支持
	return redis.String(conn.Do("XADD", command...))
}

func (c *Client) XGROUP(ctx context.Context, groupID, topic string, startIndex *int) error {
	return c.xGroup(ctx, groupID, topic, startIndex)
}

func (c *Client) xGroup(ctx context.Context, groupID, topic string, startIndex *int) error {
	if groupID == "" || topic == "" {
		return errors.New("redis xReadGroup groupID/topic can't be empty")
	}

	// 获取链接
	conn, err := c.pool.GetContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	if startIndex == nil {
		// Simple string reply: OK on success 无需校验
		_, err = conn.Do("XGROUP", "CREATE", topic, groupID, "$", "MKSTREAM")
	} else {
		_, err = conn.Do("XGROUP", "CREATE", topic, groupID, startIndex, "MKSTREAM")
	}

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) XReadGroup(ctx context.Context, groupID, consumerID, topic string, timeoutMiliSeconds int, pending bool) ([]*MsgEntity, error) {
	return c.xReadGroup(ctx, groupID, consumerID, topic, timeoutMiliSeconds, pending)
}

type MsgEntity struct {
	MsgID string
	Key   string
	Val   string
}

func (c *Client) xReadGroup(ctx context.Context, groupID, consumerID, topic string, timeoutMiliSeconds int, pending bool) ([]*MsgEntity, error) {

	if groupID == "" || consumerID == "" || topic == "" {
		return nil, errors.New("redis xReadGroup groupID/consumerID/topic can't be empty")
	}

	// 获取链接
	conn, err := c.pool.GetContext(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var reply interface{}
	if pending {
		reply, err = conn.Do("XREADGROUP", "GROUP", groupID, consumerID, "STREAMS", topic, "0")
	} else {
		reply, err = conn.Do("XREADGROUP", "GROUP", groupID, consumerID, "BLOCK", timeoutMiliSeconds, "STREAMS", topic, ">")
	}

	if err != nil {
		return nil, err
	}
	// The command returns an array of results: each element of the returned array is an array composed of a two element containing the key name and the entries reported for that key.
	// [ [topic,[[stream-id][values...]]] ....]
	replys, _ := reply.([]interface{})
	if len(replys) == 0 {
		return nil, errors.New("no message")
	}
	// topic列表
	// 获取订阅topic 只允许单订阅
	// [topic,[[stream-id][values...]]...]
	replyElement, _ := replys[0].([]interface{})
	if len(replyElement) != 2 {
		return nil, errors.New("invalid msg format")
	}

	//// 对消费到的数据进行格式化
	var msgs []*MsgEntity
	// [[stream-id][values...]]...
	rawMsgs, _ := replyElement[1].([]interface{})
	for _, rawMsg := range rawMsgs {
		_msg, _ := rawMsg.([]interface{})
		if len(_msg) != 2 {
			return nil, errors.New("invalid msg format")
		}
		// [stream-id]
		msgID := gocast.ToString(_msg[0])
		// [values...]
		msgBody, _ := _msg[1].([]interface{})
		if len(msgBody) != 2 {
			return nil, errors.New("invalid msg format")
		}
		// 暂定为key value 模式
		msgKey := msgBody[0]
		msgVal := msgBody[1]
		msgs = append(msgs, &MsgEntity{
			MsgID: msgID,
			Key:   gocast.ToString(msgKey),
			Val:   gocast.ToString(msgVal),
		})
	}

	return msgs, err
}

func (c *Client) XACK(ctx context.Context, topic, groupID, msgID string) (int, error) {
	if groupID == "" || msgID == "" || topic == "" {
		return 0, errors.New("redis xReadGroup groupID/consumerID/topic can't be empty")
	}

	// 获取链接
	conn, err := c.pool.GetContext(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	return redis.Int(conn.Do("XACK", topic, groupID, msgID))

}
