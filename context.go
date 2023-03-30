package frame

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Context 获取 frame 上下文
type Context struct {
	*gin.Context
	config       *Config
	dbClients    *DBMultiClient
	redisClients *RedisMultiClient
	*logrus.Entry
}

func (c *Context) getGormLogger() logger.Interface {
	traceID := c.GetHeader(TraceID)
	return newGormLogger(traceID, c.config).LogMode(log2gormLevel(c.config.LogLevel))
}

// WithContext return context
func (c *Context) WithContext() context.Context {
	return context.WithValue(context.Background(), TraceID, c.GetHeader(TraceID))
}

// GetDB get db client
func (ctx *Context) GetDB(name ...string) *gorm.DB {
	// default mysql client
	if len(ctx.dbClients.clients) == 1 && len(name) == 0 {
		for _, v := range ctx.dbClients.clients {
			v = v.WithContext(ctx.WithContext())
			v.Logger = ctx.getGormLogger()
			return v
		}
	}
	if len(name) == 0 {
		panic("db client can't find, db name is empty")
	}
	db := ctx.dbClients.clients[name[0]]
	if db != nil {
		db = db.WithContext(ctx.WithContext())
		db.Logger = ctx.getGormLogger()
	}
	return db
}

// GetRedis get redis client
func (ctx *Context) GetRedis(name ...string) *redis.Client {
	// default redis client
	if len(ctx.redisClients.clients) == 1 && len(name) == 0 {
		for _, v := range ctx.redisClients.clients {
			return v
		}
	}
	if len(name) == 0 {
		panic("redis client can't find, redis name is empty")
	}
	return ctx.redisClients.clients[name[0]]
}

func (ctx *Context) GetLogger() *logrus.Entry {
	traceID := ctx.GetHeader(TraceID)
	return ctx.Entry.WithField(TraceID, traceID)
}
