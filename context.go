package frame

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Context 获取 frame 上下文
type Context struct {
	*gin.Context
	db     *gin.Context
	log    *logrus.Entry
	config *Config
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
