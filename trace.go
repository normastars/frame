package frame

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"gorm.io/gorm/logger"
)

var defaultJSONLogFormatter = &logrus.JSONFormatter{
	CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
		return frame.Function, path.Base(frame.File)
	},
}

// SetDefaultLog config project logrus log
func SetDefaultLog() {
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
	if mode == "" || mode == ModeJSON {
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

		// request start
		startTime := time.Now()

		//  request body
		var requestBody string
		if c.Request.Body != nil {
			bodyBytes, err := ioutil.ReadAll(c.Request.Body)
			if err != nil {
				log.WithError(err).Error("Failed to read request body")
			} else {
				requestBody = string(bodyBytes)
			}

			// reset body
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// request header
		requestHeader := make(map[string]string)
		for k, v := range c.Request.Header {
			requestHeader[k] = strings.Join(v, ",")
		}

		// response body
		w := &responseWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Context.Writer}
		c.Context.Writer = w

		// detail request
		c.Next()

		// request end
		endTime := time.Now()
		duration := endTime.Sub(startTime).Milliseconds()
		// log body
		rb := w.body.String()
		if !isJSONBody(w) {
			return
		}
		go func() {
			httpCode := c.Context.Writer.Status()
			hcr := fmt.Sprintf("%d", httpCode)
			busCode := jsonGet(rb, codeKey)
			method := c.Context.Request.Method
			url := c.Request.URL.Path

			if c.config.EnableMetric {
				// metrics
				prometheusRequestDuration.WithLabelValues(url, hcr, method).Observe(float64(duration))
				prometheusRequestBusCounter.WithLabelValues(url, busCode, method).Inc()
			}
			if c.config.HTTPServer.DisableReqLog {
				return
			}
			reqLog := logBody{
				RequestID:  c.Request.Header.Get(TraceIDKey),
				Code:       busCode,
				StatusCode: c.Context.Writer.Status(),
				Duration:   duration,
				Msg:        jsonGet(rb, msgKey),
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
			c.Infoln(string(byts))
		}()
	}
}

func jsonGet(data string, key string) string {
	return gjson.Get(data, key).String()
}

func isJSONBody(w gin.ResponseWriter) bool {
	t := w.Header().Get("Content-Type")
	if strings.Contains(t, "application/json") {
		return true
	}
	return false
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
