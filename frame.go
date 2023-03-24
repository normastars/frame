package frame

import "github.com/gin-gonic/gin"

func Default() *Engine {
	e := &Engine{
		Engine: defaultEngine(),
		Level:  1,
		// log: nil,
	}
	return e

}

func defaultEngine() *gin.Engine {
	r := gin.Default()
	r.Use(RequestID())
	return r
}

type Engine struct {
	*gin.Engine
	Level int
}
