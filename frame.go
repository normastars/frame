package frame

import (
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/sirupsen/logrus"
)

// TODO LIST:
// 1. gin.Context -> frame.Context
// 2. 集成 Config
// 3. 优化 Redis,Mysql 连接,日志打印
// 4. 跑通demo
// 5. 集成 req 请求
// 6. 优化 Log 包

// Engine frame engine
type Engine struct {
	*gin.Engine
	DB  *gorm.DB
	Log *logrus.Entry
}

// Run engin run
func (e *Engine) Run(addr string) error {
	return e.Engine.Run(addr)
}

// Default engin
func Default() *Engine {
	// 关闭Gin的日志输出
	gin.DefaultWriter = ioutil.Discard
	e := &Engine{
		Engine: defaultEngine(),
		// DB:     newGrom("TODO"),
	}
	return e

}

func NewEngine() *Engine {
	// 初始化 logrus 日志包
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// // 初始化 gorm 数据库连接
	// db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	// if err != nil {
	// 	logger.WithError(err).Fatal("Failed to connect to database")
	// }

	return &Engine{
		Engine: defaultEngine(),
		// DB:     db,
		Log: logger.WithField("component", "MyEngine"),
	}
}

func defaultEngine() *gin.Engine {
	r := gin.Default()
	r.Use(TraceFunc())
	r.Use(LoggerFunc())
	return r
}

func (e *Engine) createContext(c *gin.Context) *Context {
	return &Context{
		// TODO: config, log, mysql, redis
		Context: c,
		// DB:      e.DB,
		// Log: e.Log.WithField("request_id", c.GetString("request_id")),
	}
}

// Use middleware
func (e *Engine) Use(middleware ...gin.HandlerFunc) {
	e.Engine.Use(middleware...)
}

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
