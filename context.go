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
	Gtx           *gin.Context
	config        *Config
	configManager *ConfigManager
	*logrus.Entry
	httpClient *req.Client
	traceID    string
	coreApp    *coreApp
	gormLogger logger.Interface
	redisHook  redis.Hook
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

// WithTraceContext return context with trace_id
func (c *Context) WithTraceContext() context.Context {
	id := c.GetTraceID()
	pc := context.Background()
	return context.WithValue(pc, TraceIDKey, id)
}

// GetDB get db client
func (c *Context) GetDB(name ...string) *gorm.DB {
	if c.coreApp == nil || c.coreApp.dbManager == nil {
		return nil
	}

	var db *gorm.DB
	if c.coreApp.mockDB != nil {
		db = c.coreApp.mockDB.DB
	} else {
		db = c.coreApp.dbManager.GetDB(name...)
	}

	if db == nil {
		return nil
	}
	db = db.WithContext(c.WithTraceContext())
	db.Logger = c.gormLogger
	return db
}

// GetRedis get redis client
func (c *Context) GetRedis(name ...string) *redis.Client {
	if c.coreApp == nil {
		return nil
	}

	if c.coreApp.mockRedis != nil {
		return c.coreApp.mockRedis
	}
	return c.coreApp.cacheManager.GetRedis(name...)
}

// GetSetTraceHeader get trace_id from header, will set trace_id in header when header trace_id is empty
func (c *Context) GetSetTraceHeader() string {
	traceID := c.GetTraceID()
	if len(traceID) > 0 {
		return traceID
	}
	traceID = generateTraceID(c.config.Project)
	c.Gtx.Header(TraceIDKey, traceID)
	c.traceID = traceID
	return traceID
}

// GetLogger get ctx log
func (c *Context) GetLogger() *logrus.Entry {
	traceID := c.GetSetTraceHeader()
	return c.Entry.WithField(TraceIDKey, traceID)
}

// GetConfig return app config
func (c *Context) GetConfig() *Config {
	return c.config
}

// GetConfigManager return config manager
func (c *Context) GetConfigManager() *ConfigManager {
	return c.configManager
}
