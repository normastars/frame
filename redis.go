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

func openRedisServers(config RedisConfig) {
	// 只会初始化一次
	redisOnce.Do(func() {
		if len(config.Configs) > 0 && config.Enable {
			for _, v := range config.Configs {
				openRedis(v)
			}
		}
		return
	})
}

func openRedis(item RedisConfigItem) {
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
