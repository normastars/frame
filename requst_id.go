package frame

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// TraceFunc trace id funcs
func TraceFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.Request.Header.Get(TraceID)
		if traceID == "" {
			traceID = uuid.NewString()
			c.Request.Header.Set(TraceID, traceID)
		}
		logger := logrus.New()
		logger.Level = logrus.DebugLevel
		logger.AddHook(NewHook(traceID))
		c.Writer.Header().Set(TraceID, traceID)
		c.Next()
	}
}
