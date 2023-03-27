package frame

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Context 获取 frame 上下文
type Context struct {
	*gin.Context
	config       *Config
	dbClients    *DBMultiClient
	redisClients *RedisMultiClient
	log          *logrus.Entry
}

// GetDB get db client
func (ctx *Context) GetDB(name ...string) *gorm.DB {
	// default mysql client
	if len(ctx.dbClients.clients) == 1 && len(name) == 0 {
		for _, v := range ctx.dbClients.clients {
			return v
		}
	}
	if len(name) == 0 {
		panic("db client can't find, db name is empty")
	}
	return ctx.dbClients.clients[name[0]]
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

// // DB 获取GORM数据库连接
// func (ctx *Context) DB() *gorm.DB {
// 	db, ok := ctx.Get("db")
// 	if !ok {
// 		panic("Database connection not found in context")
// 	}
// 	return db.(*gorm.DB)
// }

// // Redis 获取Redis连接
// func (ctx *Context) Redis() *redis.Client {
// 	r, ok := ctx.Get("redis")
// 	if !ok {
// 		panic("Redis connection not found in context")
// 	}
// 	c, _ := r.(*redis.Client)
// 	return c
// }

// // Config 获取配置
// func (ctx *Context) Config() map[string]interface{} {
// 	config, ok := ctx.Get("config")
// 	if !ok {
// 		panic("Config not found in context")
// 	}
// 	return config.(map[string]interface{})
// }

// // ContextMiddleware 中间件：将GORM数据库连接、Redis连接和配置存储到上下文中
// func ContextMiddleware(db *gorm.DB, redis *redis.Client, config map[string]interface{}) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		ctx := &Context{Context: c}
// 		ctx.Set("db", db)
// 		ctx.Set("redis", redis)
// 		ctx.Set("config", config)
// 		c.Next()
// 	}
// }

// // ToContext 将gin.Context转换为MyContext
// func ToContext(c *gin.Context) *Context {
// 	return &Context{Context: c}
// }

// // FromContext 获取上下文
// func FromContext(ctx context.Context) *Context {
// 	c, ok := ctx.(*gin.Context)
// 	if !ok {
// 		panic("Context is not of type *gin.Context")
// 	}
// 	return &Context{Context: c}
// }
