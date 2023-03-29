package frame

import (
	"sync"

	"github.com/go-redis/redis"
)

var redisOnce sync.Once

// RedisMultiClient multi db conns
type RedisMultiClient struct {
	clients map[string]*redis.Client
}

var redisMultiConn = &RedisMultiClient{
	clients: map[string]*redis.Client{},
}

// GetRedisConn 获取 redis 链接
func GetRedisConn() *RedisMultiClient {
	return redisMultiConn
}

func newRedisServers(conf *Config) {
	// 只会初始化一次
	redisOnce.Do(func() {
		if len(conf.Redis.Configs) > 0 && conf.Redis.Enable {
			for _, v := range conf.Redis.Configs {
				openRedis(v)
			}
		}
		return
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
