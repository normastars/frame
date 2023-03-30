package frame

import (
	"encoding/base64"
	"strings"

	"github.com/google/uuid"
)

// TraceFunc trace id func
func TraceFunc() HandlerFunc {
	return func(c *Context) {
		traceID := c.Request.Header.Get(TraceIDKey)
		if traceID == "" {
			traceID = generalTraceID(c.config.Project)
			c.Request.Header.Set(TraceIDKey, traceID)
		}
		c.Context.Writer.Header().Set(TraceIDKey, traceID)
		c.Next()
	}
}

func generalTraceID(project ...string) string {
	prefix := ""
	if len(project) > 0 && project[0] != "" {
		prefix = base64.StdEncoding.EncodeToString([]byte(project[0]))
		prefix = strings.ReplaceAll(prefix, "=", "-")
	}
	traceID := uuid.NewString()
	if len(prefix) > 0 {
		return prefix + traceID
	}
	return traceID
}
