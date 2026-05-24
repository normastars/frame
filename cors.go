package frame

import "net/http"

// defaultCORSAllowHeaders is the explicit header allowlist used when CORS is
// enabled. Using "*" here would expose authentication headers to any origin;
// explicit listing is safer and still covers all common SPA use-cases.
const defaultCORSAllowHeaders = "Authorization, Content-Type, Accept, X-Requested-With, " + TraceIDKey

// defaultCORSExposeHeaders lists headers the browser is allowed to read from
// the response. Only trace_id needs to be exposed for client-side correlation.
const defaultCORSExposeHeaders = TraceIDKey

// defaultCORSMaxAge is the number of seconds browsers may cache the preflight
// response (Access-Control-Max-Age). 86400 = 24 h, typical for production.
const defaultCORSMaxAge = "86400"

// CORSFunc returns a middleware that sets CORS headers.
// When allowOrigins is empty or ["*"] every origin is allowed (wildcard).
// When a specific list is provided only those origins receive the header.
func CORSFunc(allowOrigins ...string) HandlerFunc {
	allowOrigin := "*"
	if len(allowOrigins) > 0 && allowOrigins[0] != "" {
		allowOrigin = allowOrigins[0]
	}
	return func(c *Context) {
		method := c.Gtx.Request.Method
		origin := c.Gtx.Request.Header.Get("Origin")
		if origin != "" {
			if allowOrigin == "*" {
				c.Gtx.Header("Access-Control-Allow-Origin", "*")
			} else {
				// Only echo the origin back when it matches the allowlist.
				for _, o := range allowOrigins {
					if o == origin {
						c.Gtx.Header("Access-Control-Allow-Origin", origin)
						// Vary tells caches that the response differs per origin.
						c.Gtx.Header("Vary", "Origin")
						break
					}
				}
			}
			c.Gtx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Gtx.Header("Access-Control-Allow-Headers", defaultCORSAllowHeaders)
			c.Gtx.Header("Access-Control-Expose-Headers", defaultCORSExposeHeaders)
			c.Gtx.Header("Access-Control-Max-Age", defaultCORSMaxAge)
		}
		if method == http.MethodOptions {
			c.Gtx.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Gtx.Next()
	}
}
