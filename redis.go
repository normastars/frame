package frame

import (
	"context"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// RedisMultiClient multi db conns
type RedisMultiClient struct {
	clients map[string]*redis.Client
}

// GetRedisConn returns the redis client map (backward-compatible function)
func GetRedisConn() *RedisMultiClient {
	core := getActiveCore()
	if core == nil {
		return &RedisMultiClient{clients: make(map[string]*redis.Client)}
	}
	return core.legacyRedisConns
}

// redisLogHook is a Redis log hook
type redisLogHook struct {
	Log     *logrus.Logger
	Disable bool
}

func newRedisLogHook(config *Config) redis.Hook {
	return &redisLogHook{Log: NewLogger(config), Disable: config.Redis.DisableReqLog}
}

func (l redisLogHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	if l.Disable {
		return ctx, nil
	}
	l.Log.WithFields(logrus.Fields{
		TraceIDKey: getTraceIDFromContext(ctx),
	}).Infof("Redis command: %s", cmd.String())
	return ctx, nil
}

func (l redisLogHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	return nil
}

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

func (l redisLogHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	return nil
}
