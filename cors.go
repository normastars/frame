package frame

import "net/http"

// CORSFunc cors middleware
func CORSFunc() HandlerFunc {
	return func(c *Context) {
		method := c.Gtx.Request.Method
		origin := c.Gtx.Request.Header.Get("Origin")
		if origin != "" {
			c.Gtx.Header("Access-Control-Allow-Origin", "*") // 可将将 * 替换为指定的域名
			c.Gtx.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Gtx.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
			c.Gtx.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
			c.Gtx.Header("Access-Control-Allow-Credentials", "true")
		}
		if method == "OPTIONS" {
			c.Gtx.AbortWithStatus(http.StatusNoContent)
		}
		c.Gtx.Next()
	}
}
