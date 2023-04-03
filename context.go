package frame

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/imroc/req/v3"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Context frame context
type Context struct {
	Gtx          *gin.Context
	config       *Config
	dbClients    *DBMultiClient
	redisClients *RedisMultiClient
	*logrus.Entry
	httpClient *req.Client
	traceID    string
}

// GetTraceID return trace id from context
func (c *Context) GetTraceID() string {
	if c.Gtx == nil {
		if c.traceID != "" {
			return c.traceID
		}
		return ""
	}
	return c.Gtx.GetHeader(TraceIDKey)
}

// DoHTTP return http client
func (c *Context) DoHTTP() *req.Client {
	return c.httpClient
}

func (c *Context) getGormLogger() logger.Interface {
	return newGormLogger(c.config).LogMode(log2gormLevel(c.config.LogLevel))
}

// WithTraceContext return context
func (c *Context) WithTraceContext() context.Context {
	id := c.GetTraceID()
	pc := context.Background()
	return context.WithValue(pc, TraceIDKey, id)
}

// GetDB get db client
func (c *Context) GetDB(name ...string) *gorm.DB {
	// default mysql client
	if len(c.dbClients.clients) == 1 && len(name) == 0 {
		for _, v := range c.dbClients.clients {
			v = v.WithContext(c.WithTraceContext())
			v.Logger = c.getGormLogger()
			return v
		}
	}
	// panic when db
	if len(name) == 0 {
		panic("db client can't find, db name is empty")
	}
	db := c.dbClients.clients[name[0]]
	if db != nil {
		db = db.WithContext(c.WithTraceContext())
		db.Logger = c.getGormLogger()
	}
	return db
}

// GetRedis get redis client
func (c *Context) GetRedis(name ...string) *redis.Client {
	// default redis client
	if len(c.redisClients.clients) == 1 && len(name) == 0 {
		for _, v := range c.redisClients.clients {
			v = v.WithContext(c.WithTraceContext())
			v.AddHook(newRedisLogHook(c.config))
			return v
		}
	}
	if len(name) == 0 {
		panic("redis client can't find, redis name is empty")
	}
	r := c.redisClients.clients[name[0]]
	r = r.WithContext(c.WithTraceContext())
	r.AddHook(newRedisLogHook(c.config))
	return r
}

// GetSetTraceHeader get trace_id from header, will set trace_id in header when header trace_id is empty
func (c *Context) GetSetTraceHeader() string {
	traceID := c.GetTraceID()
	if len(traceID) > 0 {
		return traceID
	}
	traceID = generalTraceID(c.config.Project)
	c.Gtx.Header(TraceIDKey, traceID)
	c.traceID = traceID
	return traceID
}

// GetLogger get ctx log
func (c *Context) GetLogger() *logrus.Entry {
	traceID := c.GetSetTraceHeader()
	return c.Entry.WithField(TraceIDKey, traceID)
}
