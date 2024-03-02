package frame

import (
	"encoding/base64"
	"strings"

	"github.com/google/uuid"
)

// TraceFunc trace id func
func TraceFunc() HandlerFunc {
	return func(c *Context) {
		traceID := c.Gtx.Request.Header.Get(TraceIDKey)
		if traceID == "" {
			traceID = generateTraceID(c.config.Project)
			c.Gtx.Request.Header.Set(TraceIDKey, traceID)
		}
		c.traceID = traceID
		c.Gtx.Writer.Header().Set(TraceIDKey, traceID)
		c.Gtx.Next()
	}
}

func generateTraceID(project ...string) string {
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
