package frame

import (
	"bytes"
	"io/ioutil"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间
		startTime := time.Now()
		// 请求body内容处理
		var requestBody string
		if c.Request.Body != nil {
			bodyBytes, err := ioutil.ReadAll(c.Request.Body)
			if err != nil {
				log.WithError(err).Error("Failed to read request body")
			} else {
				requestBody = string(bodyBytes)
			}
			// 重置Body内容
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// 处理请求header
		requestHeader := make(map[string]string)
		for k, v := range c.Request.Header {
			requestHeader[k] = strings.Join(v, ",")
		}

		// 处理响应body内容
		w := &responseWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = w

		// 继续处理请求
		c.Next()

		// 记录请求结束时间
		endTime := time.Now()

		// 日志记录
		log.WithFields(log.Fields{
			"requestId":   c.Request.Header.Get("X-Request-ID"),
			"requestUrl":  c.Request.URL.Path,
			"requestBody": requestBody,
			"requestHdr":  requestHeader,
			"statusCode":  c.Writer.Status(),
			"response":    w.body.String(),
			"latency":     endTime.Sub(startTime),
		}).Info("HTTP request completed")
	}
}

type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
