package frame

import "github.com/gin-gonic/gin"

// HandlerFunc freme middleware
type HandlerFunc func(*Context)

// Use use middleware
func (e *Engine) Use(middleware ...HandlerFunc) {
	if len(middleware) > 0 {
		for i := range middleware {
			e.Engine.Use(e.convert2GinHandlerFunc(middleware[i]))
		}
	}
}

func (e *Engine) convert2FrameContext(c *gin.Context) *Context {
	// set log trace_id
	traceID := c.GetHeader(TraceID)
	l := e.log.WithField(TraceID, traceID)
	return &Context{
		Context:      c,
		config:       e.config,
		dbClients:    e.dbClients,
		redisClients: e.redisClients,
		Entry:        l,
	}
}

func (e *Engine) convert2GinHandlerFunc(h HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := e.convert2FrameContext(c)
		h(ctx)
	}
}
