package frame

import "github.com/gin-gonic/gin"

// RouterGroup wraps gin.RouterGroup to provide frame-compatible route
// registration with full Context support (config, DB, Redis, HTTP client,
// logging). Use App.Group() to create a RouterGroup.
type RouterGroup struct {
	app *App
	rg  *gin.RouterGroup
}

// BasePath returns the base path of the router group.
func (rg *RouterGroup) BasePath() string {
	return rg.rg.BasePath()
}

// Use registers frame-compatible middleware on the group.
func (rg *RouterGroup) Use(middleware ...HandlerFunc) {
	for _, m := range middleware {
		rg.rg.Use(rg.app.convert2GinHandlerFunc(m))
	}
}

// Group creates a nested router group with optional middleware.
func (rg *RouterGroup) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	ginHandlers := make([]gin.HandlerFunc, len(handlers))
	for i, h := range handlers {
		ginHandlers[i] = rg.app.convert2GinHandlerFunc(h)
	}
	return &RouterGroup{
		app: rg.app,
		rg:  rg.rg.Group(relativePath, ginHandlers...),
	}
}

// GET registers a GET route on the group.
func (rg *RouterGroup) GET(relativePath string, handler func(c *Context)) {
	rg.rg.GET(relativePath, rg.app.registerRouteWithContext(handler))
}

// POST registers a POST route on the group.
func (rg *RouterGroup) POST(relativePath string, handler func(c *Context)) {
	rg.rg.POST(relativePath, rg.app.registerRouteWithContext(handler))
}

// PUT registers a PUT route on the group.
func (rg *RouterGroup) PUT(relativePath string, handler func(c *Context)) {
	rg.rg.PUT(relativePath, rg.app.registerRouteWithContext(handler))
}

// PATCH registers a PATCH route on the group.
func (rg *RouterGroup) PATCH(relativePath string, handler func(c *Context)) {
	rg.rg.PATCH(relativePath, rg.app.registerRouteWithContext(handler))
}

// DELETE registers a DELETE route on the group.
func (rg *RouterGroup) DELETE(relativePath string, handler func(c *Context)) {
	rg.rg.DELETE(relativePath, rg.app.registerRouteWithContext(handler))
}

// HEAD registers a HEAD route on the group.
func (rg *RouterGroup) HEAD(relativePath string, handler func(c *Context)) {
	rg.rg.HEAD(relativePath, rg.app.registerRouteWithContext(handler))
}
