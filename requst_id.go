package frame

import (
	"github.com/google/uuid"
)

// TraceFunc trace id funcs
func TraceFunc() HandlerFunc {
	return func(c *Context) {
		traceID := c.Request.Header.Get(TraceID)
		if traceID == "" {
			traceID = uuid.NewString()
			c.Request.Header.Set(TraceID, traceID)
		}
		c.Context.Writer.Header().Set(TraceID, traceID)
		// c.Writer.Header().Set(TraceID, traceID)
		c.Next()
	}
}
