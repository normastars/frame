package frame

import (
	"io/ioutil"

	"github.com/gin-gonic/gin"
)

// Default engin
func Default() *Engine {
	// 关闭Gin的日志输出
	gin.DefaultWriter = ioutil.Discard
	e := &Engine{
		Engine: defaultEngine(),
		Level:  "INFO",
	}
	return e

}

func defaultEngine() *gin.Engine {
	r := gin.Default()
	r.Use(TraceFunc())
	r.Use(LoggerFunc())
	return r
}

// Engine frame engine
type Engine struct {
	*gin.Engine
	Level string
}
