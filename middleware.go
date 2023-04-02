package frame

import "github.com/gin-gonic/gin"

// HandlerFunc freme middleware
type HandlerFunc func(*Context)

// Use use middleware
func (e *App) Use(middleware ...HandlerFunc) {
	if len(middleware) > 0 {
		for i := range middleware {
			e.Engine.Use(e.convert2GinHandlerFunc(middleware[i]))
		}
	}
}

func (e *App) convert2FrameContext(c *gin.Context) *Context {
	return &Context{
		Gtx:          c,
		config:       e.config,
		dbClients:    e.dbClients,
		redisClients: e.redisClients,
		Entry:        e.getLogEntry(c),
		httpClient:   e.getHTTPClient(c),
	}
}

func (e *App) convert2GinHandlerFunc(h HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := e.convert2FrameContext(c)
		h(ctx)
	}
}
