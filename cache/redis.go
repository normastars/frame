// Package cache provides Redis connection management.
// Instance-level Manager with lazy initialization, eliminating global state.
package cache

import (
	"context"
	"errors"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/normastars/frame/internal"
	"github.com/sirupsen/logrus"
)

// Manager manages Redis connections (instance-level, concurrency-safe)
type Manager struct {
	configs []internal.RedisConfig
	clients map[string]*redis.Client
	mu      sync.RWMutex
	once    sync.Once
	initErr error
}

// NewManager creates a new Redis connection manager.
// Does not connect immediately; initializes lazily on first GetRedis call.
func NewManager(configs []internal.RedisConfig) *Manager {
	return &Manager{
		configs: configs,
		clients: make(map[string]*redis.Client),
	}
}

// GetRedis returns a Redis client. Lazy initialized on first call.
func (m *Manager) GetRedis(name ...string) *redis.Client {
	m.initConnections()
	if m.initErr != nil {
		logrus.Errorf("cache: init error: %v", m.initErr)
		return nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.clients) == 1 && len(name) == 0 {
		for _, v := range m.clients {
			return v
		}
	}
	if len(name) == 0 {
		logrus.Error("cache: GetRedis called with empty name but multiple clients exist")
		return nil
	}
	return m.clients[name[0]]
}

// initConnections lazy initializes all Redis connections.
// The closure runs exclusively inside once.Do, so no additional locking is
// needed while writing to m.clients — the RWMutex is only needed for concurrent
// reads that happen after initialization completes.
func (m *Manager) initConnections() {
	m.once.Do(func() {
		if len(m.configs) == 0 {
			return
		}
		var errs []error
		for _, v := range m.configs {
			if !v.Enable {
				continue
			}
			client, err := m.open(v)
			if err != nil {
				logrus.Errorf("cache: failed to connect redis %s: %v", v.Name, err)
				errs = append(errs, err)
				continue
			}
			m.clients[v.Name] = client
		}
		if len(errs) > 0 {
			m.initErr = errors.Join(errs...)
		}
	})
}

func (m *Manager) open(item internal.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     item.Host,
		Password: item.Password,
		PoolSize: item.PoolSize,
		DB:       item.DB,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		client.Close()
		return nil, err
	}
	return client, nil
}
