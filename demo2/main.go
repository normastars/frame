package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type MyEngine struct {
	*gin.Engine
	// DB  *gorm.DB
	Log *logrus.Entry
}

type MyContext struct {
	*gin.Context
	// DB  *gorm.DB
	Log *logrus.Entry
}

func (e *MyEngine) Run(addr string) error {
	return e.Engine.Run(addr)
}

func NewMyEngine() *MyEngine {
	engine := gin.New()

	// 初始化 logrus 日志包
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// // 初始化 gorm 数据库连接
	// db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	// if err != nil {
	// 	logger.WithError(err).Fatal("Failed to connect to database")
	// }

	return &MyEngine{
		Engine: engine,
		// DB:     db,
		Log: logger.WithField("component", "MyEngine"),
	}
}

func (e *MyEngine) createContext(c *gin.Context) *MyContext {
	return &MyContext{
		Context: c,
		// DB:      e.DB,
		Log: e.Log.WithField("request_id", c.GetString("request_id")),
	}
}

func (e *MyEngine) Use(middleware ...gin.HandlerFunc) {
	e.Engine.Use(middleware...)
}

func (e *MyEngine) GET(relativePath string, handler func(c *MyContext)) {
	e.Engine.GET(relativePath, func(c *gin.Context) {
		ctx := e.createContext(c)

		// 设置 request_id，用于日志追踪
		requestID := c.Request.Header.Get("X-Request-ID")
		if requestID != "" {
			ctx.Set("request_id", requestID)
		}
		handler(ctx)
	})
}

func main() {
	engine := NewMyEngine()

	// 注册 /hello 接口
	engine.GET("/hello", func(c *MyContext) {
		// 使用 MyContext 的 Log 和 JSON 函数
		c.Log.Debug("Received request")
		c.JSON(http.StatusOK, gin.H{"message": "hello world"})
	})

	// 启动 HTTP 服务器
	err := engine.Run(":8080")
	if err != nil {
		engine.Log.WithError(err).Fatal("Failed to start HTTP server")
	}
}
