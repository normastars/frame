package frame

import "net/http"

// CORSFunc cors middleware
func CORSFunc() HandlerFunc {
	return func(c *Context) {
		method := c.Gtx.Request.Method
		origin := c.Gtx.Request.Header.Get("Origin")
		if origin != "" {
			c.Gtx.Header("Access-Control-Allow-Origin", "*")
			c.Gtx.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Gtx.Header("Access-Control-Allow-Headers", "*")
			c.Gtx.Header("Access-Control-Expose-Headers", "*")
		}
		if method == "OPTIONS" {
			c.Gtx.AbortWithStatus(http.StatusNoContent)
		}
		c.Gtx.Next()
	}
}
