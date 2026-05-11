package frame

import (
	"github.com/go-redis/redis/v8"
	"github.com/normastars/frame/core"
	"gorm.io/gorm"
)

// Option is a framework configuration option. The only new API visible to external callers.
// Usage:
//
//	app := frame.New(frame.WithMockDB(sqliteDB))
type Option = core.Option

// WithMockDB injects a mock database connection for unit tests.
// Use an in-memory SQLite database or other mock implementations.
func WithMockDB(db *gorm.DB) Option {
	return core.WithMockDB(db)
}

// WithMockRedis injects a mock Redis client for unit tests.
func WithMockRedis(client *redis.Client) Option {
	return core.WithMockRedis(client)
}
