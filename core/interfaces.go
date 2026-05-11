// Package core provides core internal interfaces and data types for the frame framework.
// All interfaces are unexported (lowercase). External users only need to import the frame root package.
package core

import "gorm.io/gorm"

// DBConfig holds MySQL connection config (internal, converted from root Config)
type DBConfig struct {
	Name              string
	Enable            bool
	EnableAutoMigrate bool
	Host              string
	Database          string
	User              string
	Password          string
	SlowThresholdSec  int
	DisableReqLog     bool
}

// RedisConfig holds Redis connection config (internal)
type RedisConfig struct {
	Name          string
	Enable        bool
	Host          string
	PoolSize      int
	Password      string
	DB            int
	DisableReqLog bool
}

// MockDBHolder holds the mock DB for WithMockDB option
type MockDBHolder struct {
	DB *gorm.DB
}
