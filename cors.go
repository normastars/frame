package frame

import "net/http"

// CORSFunc cors middleware
func CORSFunc(allowOrigins ...string) HandlerFunc {
	allowOrigin := "*"
	if len(allowOrigins) > 0 && allowOrigins[0] != "" {
		allowOrigin = allowOrigins[0]
	}
	return func(c *Context) {
		method := c.Gtx.Request.Method
		origin := c.Gtx.Request.Header.Get("Origin")
		if origin != "" {
			if allowOrigin != "*" {
				// only allow specific origin
				for _, o := range allowOrigins {
					if o == origin {
						c.Gtx.Header("Access-Control-Allow-Origin", origin)
						break
					}
				}
			} else {
				c.Gtx.Header("Access-Control-Allow-Origin", "*")
			}
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
