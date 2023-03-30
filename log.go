package frame

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm/logger"
)

var defaultJSONLogFormatter = &logrus.JSONFormatter{
	CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
		return frame.Function, path.Base(frame.File)
	},
}

// init config project logrus log
func init() {
	// default log config
	logrus.SetReportCaller(true)
	// set file name
	logrus.SetFormatter(defaultJSONLogFormatter)
}

// NewLogger new logger
func NewLogger(conf ...*Config) *logrus.Logger {
	var l string
	var m string
	if len(conf) > 0 {
		l = conf[0].LogLevel
		m = conf[0].LogMode
	}
	return newLoggerLevel(l, m)
}

func newLoggerLevel(level, mode string) *logrus.Logger {
	logger := logrus.New()
	logger.SetReportCaller(true)
	if mode == "" || mode == "json" {
		logger.SetFormatter(defaultJSONLogFormatter)
	}
	if len(level) <= 0 {
		logger.SetLevel(logrus.DebugLevel)
		return logger
	}
	// set log level, default info level
	logger.SetLevel(log2Level(level))
	return logger
}

var logm = map[string]logrus.Level{
	"panic": logrus.PanicLevel,
	"fatal": logrus.FatalLevel,
	"error": logrus.ErrorLevel,
	"warn":  logrus.WarnLevel,
	"info":  logrus.InfoLevel,
	"debug": logrus.DebugLevel,
	"trace": logrus.TraceLevel,
}

var gormLogm = map[string]logger.LogLevel{
	"error": logger.Error,
	"warn":  logger.Warn,
	"info":  logger.Info,
}

func log2gormLevel(l string) logger.LogLevel {
	l = strings.ToLower(l)
	le, ok := gormLogm[l]
	if ok {
		return le
	}
	// default silent level
	return logger.Silent
}

func log2Level(l string) logrus.Level {
	l = strings.ToLower(l)
	le, ok := logm[l]
	if ok {
		return le
	}
	// default info level
	return logrus.InfoLevel
}

// LoggerFunc log func
func LoggerFunc() HandlerFunc {
	return func(c *Context) {
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
		w := &responseWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Context.Writer}
		c.Context.Writer = w

		// 继续处理请求
		c.Next()

		// 记录请求结束时间
		endTime := time.Now()
		// 日志记录
		reqLog := logBody{
			RequestID:  c.Request.Header.Get(TraceID),
			Code:       "X001",
			StatusCode: c.Context.Writer.Status(),
			Duration:   endTime.Sub(startTime).Milliseconds(),
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
		c.Errorln("richardyu")
		c.Infoln(string(byts))
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
	Duration   int64       `json:"duration,omitempty"` // ms
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
