package frame

import (
	"context"
	"strings"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

var redisOnce sync.Once

// RedisMultiClient multi db conns
type RedisMultiClient struct {
	clients map[string]*redis.Client
}

var redisMultiConn = &RedisMultiClient{
	clients: map[string]*redis.Client{},
}

// GetRedisConn return  redis client
func GetRedisConn() *RedisMultiClient {
	return redisMultiConn
}

func newRedisServers(conf *Config) {
	redisOnce.Do(func() {
		if len(conf.Redis.Configs) > 0 && conf.Redis.Enable {
			for _, v := range conf.Redis.Configs {
				openRedis(v)
			}
		}
	})
}

func openRedis(item RedisConfigItem) {
	if !item.Enable {
		return
	}
	client := redis.NewClient(&redis.Options{
		Addr:     item.Host,
		Password: item.Password,
		PoolSize: item.PoolSize,
		DB:       item.DB,
	})
	if client != nil {
		redisMultiConn.clients[item.Name] = client
	}
}

// Define a custom logging hook
type redisLogHook struct {
	Log     *logrus.Logger
	Disable bool
}

func newRedisLogHook(config *Config) redis.Hook {
	return &redisLogHook{Log: NewLogger(config), Disable: config.Redis.DisableReqLog}
}

// BeforeProcess logs the command before it is processed
func (l redisLogHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	if l.Disable {
		return ctx, nil
	}
	l.Log.WithFields(logrus.Fields{
		TraceIDKey: getTraceIDFromContext(ctx),
	}).Infof("Redis command: %s", cmd.String())
	return ctx, nil
}

// AfterProcess does nothing in this example
func (l redisLogHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	return nil
}

// BeforeProcessPipeline logs the commands before they are processed in a pipeline
func (l redisLogHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	if l.Disable {
		return ctx, nil
	}
	cmdstr := []string{}
	for _, cmd := range cmds {
		cmdstr = append(cmdstr, cmd.String())
	}
	if len(cmds) <= 0 {
		return ctx, nil
	}
	l.Log.WithFields(logrus.Fields{
		TraceIDKey: getTraceIDFromContext(ctx),
	}).Infof("Redis pipeline commands: %s", strings.Join(cmdstr, " "))
	return ctx, nil
}

// AfterProcessPipeline does nothing in this example
func (l redisLogHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	return nil
}
