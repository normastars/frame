package core

import (
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// Option is a framework configuration option.
// Usage: frame.New(frame.WithMockDB(db))
type Option func(*Options)

// Options holds internal configuration options
type Options struct {
	MockDB    *MockDBHolder
	MockRedis *redis.Client
}

// WithMockDB injects a mock database connection for unit tests.
// Use an in-memory SQLite database or other mock implementations.
func WithMockDB(db *gorm.DB) Option {
	return func(o *Options) {
		o.MockDB = &MockDBHolder{DB: db}
	}
}

// WithMockRedis injects a mock Redis client for unit tests.
func WithMockRedis(client *redis.Client) Option {
	return func(o *Options) {
		o.MockRedis = client
	}
}
