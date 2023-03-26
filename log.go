package frame

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

type Hook struct {
	TraceID string
}

func NewHook(trace_id string) *Hook {
	return &Hook{TraceID: trace_id}
}

func (h *Hook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *Hook) Fire(entry *logrus.Entry) error {
	entry.Data[TraceID] = h.TraceID
	return nil
}

// LoggerFunc log func
func LoggerFunc() gin.HandlerFunc {
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
		reqLog := logBody{
			RequestID:  c.Request.Header.Get(TraceID),
			Code:       "X001",
			StatusCode: c.Writer.Status(),
			Latency:    endTime.Sub(startTime).Microseconds(),
			Msg:        "XTODO",
			Path:       c.Request.URL.Path,
			Extra: reqLogExtra{
				Req: reqLogBody{
					// Header: requestHeader,
					Body: requestBody,
				},
				Resp: respLogBody{
					Body: w.body.String(),
				},
			},
		}
		byts, _ := json.Marshal(reqLog)

		log.Errorln("richardyu")
		log.Infoln(string(byts))
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

type logBody struct {
	RequestID  string      `json:"request_id,omitempty"`
	Code       string      `json:"code,omitempty"`
	StatusCode int         `json:"statusCode,omitempty"`
	Latency    int64       `json:"latency,omitempty"` // ms
	Msg        string      `json:"msg,omitempty"`
	Path       string      `json:"path,omitempty"`
	Extra      reqLogExtra `json:"extra,omitempty"`
}

type reqLogExtra struct {
	Req  reqLogBody  `json:"req,omitempty"`
	Resp respLogBody `json:"resp,omitempty"`
}

type reqLogBody struct {
	Header interface{} `json:"header,omitempty"`
	// Param  map[string]interface{} `json:"param,omitempty"`
	Body string `json:"body,omitempty"`
}

type respLogBody struct {
	Header interface{} `json:"header,omitempty"`
	Body   string      `json:"body,omitempty"`
}
