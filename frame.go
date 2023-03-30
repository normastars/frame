package frame

import (
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// TODO LIST:
// 1. gin.Context -> frame.Context 完成
// 4. 跑通demo  完成
// 2. 集成 Config 完成
// 3. 优化 Redis,Mysql 连接,日志打印
// 5. 集成 req 请求
// 6. 优化 Log 包

// Engine frame engine
type Engine struct {
	*gin.Engine
	config       *Config
	dbClients    *DBMultiClient
	redisClients *RedisMultiClient
	log          *logrus.Logger
}

// Run engin run
// ":8080"
func (e *Engine) Run(addr string) error {
	return e.Engine.Run(addr)
}

// Default engin
func Default() *Engine {
	// 关闭Gin的日志输出
	gin.DefaultWriter = ioutil.Discard
	e := NewEngine()
	return e
}

// NewEngine new engine
func NewEngine() *Engine {

	// step 1:  config
	conf := NewConfig()

	// step 2:  log
	logger := NewLogger(conf)

	// step 3: mysql
	newMySQLServers(conf)
	mysqlConns := GetMySQLConn()

	// step 4: redis
	newRedisServers(conf)
	redisConns := GetRedisConn()

	// 初始化 logrus 日志包

	e := &Engine{
		Engine:       defaultEngine(),
		log:          logger,
		config:       conf,
		dbClients:    mysqlConns,
		redisClients: redisConns,
	}
	e.Use(TraceFunc())
	e.Use(LoggerFunc())
	return e
}

func defaultEngine() *gin.Engine {
	r := gin.Default()
	// r.Use(TraceFunc())
	// r.Use(LoggerFunc())
	return r
}

func (e *Engine) createContext(c *gin.Context) *Context {
	// set log trace_id
	traceID := c.GetHeader(TraceID)
	l := e.log.WithField(TraceID, traceID)
	return &Context{
		Context:      c,
		config:       e.config,
		redisClients: e.redisClients,
		dbClients:    e.dbClients,
		Entry:        l,
	}
}

// Use middleware
// func (e *Engine) Use(middleware ...gin.HandlerFunc) {
// 	e.Engine.Use(middleware...)
// }

// GET get method
func (e *Engine) GET(relativePath string, handler func(c *Context)) {
	e.Engine.GET(relativePath, func(c *gin.Context) {
		ctx := e.createContext(c)
		handler(ctx)
	})
}

// POST post method
func (e *Engine) POST(relativePath string, handler func(c *Context)) {
	e.Engine.POST(relativePath, func(c *gin.Context) {
		ctx := e.createContext(c)
		handler(ctx)
	})
}

// PUT http put method
func (e *Engine) PUT(relativePath string, handler func(c *Context)) {
	e.Engine.PUT(relativePath, func(c *gin.Context) {
		ctx := e.createContext(c)
		handler(ctx)
	})
}

// PATCH  http patch method
func (e *Engine) PATCH(relativePath string, handler func(c *Context)) {
	e.Engine.PATCH(relativePath, func(c *gin.Context) {
		ctx := e.createContext(c)
		handler(ctx)
	})
}

// DELETE http delete method
func (e *Engine) DELETE(relativePath string, handler func(c *Context)) {
	e.Engine.DELETE(relativePath, func(c *gin.Context) {
		ctx := e.createContext(c)
		handler(ctx)
	})
}

// HEAD http head method
func (e *Engine) HEAD(relativePath string, handler func(c *Context)) {
	e.Engine.HEAD(relativePath, func(c *gin.Context) {
		ctx := e.createContext(c)
		handler(ctx)
	})
}
