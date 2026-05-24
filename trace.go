package frame

import (
	"bytes"
	"io"
	"net/http"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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
	logrus.SetReportCaller(true)
	logrus.SetFormatter(defaultJSONLogFormatter)
}

// NewLogger new logger
func NewLogger(conf ...*Config) *logrus.Logger {
	var level, mode string
	if len(conf) > 0 {
		level = conf[0].LogLevel
		mode = conf[0].LogMode
	}
	return newLoggerLevel(level, mode)
}

func newLoggerLevel(level, mode string) *logrus.Logger {
	logger := logrus.New()
	logger.SetReportCaller(true)
	if mode == "" || mode == ModeJSON {
		logger.SetFormatter(defaultJSONLogFormatter)
	}
	if level == "" {
		logger.SetLevel(logrus.DebugLevel)
		return logger
	}
	logger.SetLevel(log2Level(level))
	return logger
}

func log2gormLevel(l string) logger.LogLevel {
	l = strings.ToLower(l)
	le, ok := gormLogm[l]
	if ok {
		return le
	}
	return logger.Silent
}

func log2Level(l string) logrus.Level {
	l = strings.ToLower(l)
	le, ok := logm[l]
	if ok {
		return le
	}
	return logrus.InfoLevel
}

// isFileUpload checks if the request is a file upload.
// Uses HasPrefix to handle Content-Type with boundary (e.g. multipart/form-data; boundary=---).
func isFileUpload(r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")
	return strings.HasPrefix(contentType, "multipart/form-data")
}

// readRequestBody reads the full request body and restores it for downstream use.
// Returns empty string for file uploads to avoid buffering large payloads.
func readRequestBody(r *http.Request) string {
	if r.Body == nil || isFileUpload(r) {
		return ""
	}
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		logrus.Errorf("failed to read request body: %v", err)
		return ""
	}
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	return string(bodyBytes)
}

// LoggerFunc log func
func LoggerFunc() HandlerFunc {
	return func(c *Context) {
		startTime := time.Now()

		requestBody := readRequestBody(c.Gtx.Request)

		// capture values before goroutine to avoid race on c.Gtx.Request
		traceID := c.Gtx.Request.Header.Get(TraceIDKey)
		method := c.Gtx.Request.Method
		url := c.Gtx.Request.URL.Path

		// intercept response body
		w := &responseWriter{
			body:           bytes.NewBufferString(""),
			ResponseWriter: c.Gtx.Writer,
		}
		c.Gtx.Writer = w

		c.Gtx.Next()

		endTime := time.Now()
		durationMs := endTime.Sub(startTime).Milliseconds()

		if !isJSONBody(w) {
			return
		}

		// decide whether we need the goroutine at all
		enableMetric := c.config.EnableMetric
		disableLog := c.config.HTTPServer.DisableReqLog
		if !enableMetric && disableLog {
			return
		}

		responseBody := w.body.String()
		httpCode := c.Gtx.Writer.Status()

		go func() {
			defer func() {
				if r := recover(); r != nil {
					logrus.Errorf("LoggerFunc metrics/log panic: %v", r)
				}
			}()

			statusStr := strconv.Itoa(httpCode)

			if enableMetric {
				busCode := gjson.Get(responseBody, codeKey).String()
				prometheusRequestDuration.WithLabelValues(url, statusStr, method).Observe(float64(durationMs) / 1000)
				prometheusRequestBusCounter.WithLabelValues(url, busCode, method).Inc()
			}

			if disableLog {
				return
			}

			busCode := gjson.Get(responseBody, codeKey).String()
			msg := gjson.Get(responseBody, msgKey).String()

			reqLog := logBody{
				TraceType:  TraceLogRouter,
				TraceID:    traceID,
				Code:       busCode,
				StatusCode: httpCode,
				Duration:   durationMs,
				Msg:        msg,
				Path:       url,
				Extra: reqLogExtra{
					Req: reqLogBody{
						QueryParams: c.Gtx.Request.URL.Query(),
						PathParams:  c.Gtx.Params,
						Body:        requestBody,
					},
					Resp: respLogBody{
						Body: responseBody,
					},
				},
			}
			c.WithField(TraceLogKey, reqLog).Info("")
		}()
	}
}

func jsonGet(data string, key string) string {
	return gjson.Get(data, key).String()
}

func isJSONBody(w gin.ResponseWriter) bool {
	t := w.Header().Get("Content-Type")
	return strings.Contains(t, "application/json")
}

// responseWriter intercepts the response body for logging.
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

type logBody struct {
	TraceType  TraceLogType `json:"trace_type,omitempty"`
	TraceID    string       `json:"trace_id,omitempty"`
	Code       string       `json:"code,omitempty"`
	StatusCode int          `json:"status_code,omitempty"`
	Duration   int64        `json:"duration,omitempty"` // ms
	Msg        string       `json:"msg,omitempty"`
	Host       string       `json:"host,omitempty"`
	Path       string       `json:"path,omitempty"`
	Extra      reqLogExtra  `json:"extra,omitempty"`
}

type reqLogExtra struct {
	Req  reqLogBody  `json:"req,omitempty"`
	Resp respLogBody `json:"resp,omitempty"`
}

type reqLogBody struct {
	Header      http.Header         `json:"header,omitempty"`
	PathParams  interface{}         `json:"path_params,omitempty"`
	QueryParams map[string][]string `json:"query_params,omitempty"`
	Body        string              `json:"body,omitempty"`
}

type respLogBody struct {
	Body string `json:"body,omitempty"`
}
