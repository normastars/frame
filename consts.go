package frame

import (
	"github.com/sirupsen/logrus"
	"gorm.io/gorm/logger"
)

// frame consts list
const (
	TraceIDKey  = "trace_id"
	ModeJSON    = "json"
	ModeText    = "text" // was ModelText — kept consistent with ModeJSON
	TraceLogKey = "req_msg"
)

const (
	logLevelTrace = "trace"
	logLevelDebug = "debug"
	logLevelInfo  = "info"
	logLevelWarn  = "warn"
	logLevelError = "error"
	logLevelFatal = "fatal"
	logLevelPanic = "panic"
)

const (
	defaultMetricName  = "metric"
	defaultMetricsName = "metrics"
	defaultMetricPath  = "/metrics"
	defaultMetricPort  = ":9090"
	defaultMetricPort2 = "9090"
)

var logm = map[string]logrus.Level{
	logLevelPanic: logrus.PanicLevel,
	logLevelFatal: logrus.FatalLevel,
	logLevelError: logrus.ErrorLevel,
	logLevelWarn:  logrus.WarnLevel,
	logLevelInfo:  logrus.InfoLevel,
	logLevelDebug: logrus.DebugLevel,
	logLevelTrace: logrus.TraceLevel,
}

var gormLogm = map[string]logger.LogLevel{
	logLevelError: logger.Error,
	logLevelWarn:  logger.Warn,
	logLevelInfo:  logger.Info,
}

const (
	codeKey     = "code"
	msgKey      = "message"
	successMsg  = "ok"
	successCode = "0"
)

var defaultEmptyPage = PageResults{
	Results: make([]interface{}, 0),
}

// config type
const (
	configTypeYaml    = "yaml"
	configTypeYal     = "yml"
	configTypeJSON    = "json" // default json
	configPath        = "CONFPATH"
	configDefaultPath = "./conf/default.json"
)

// TraceLogType trace log type
type TraceLogType string

const (
	TraceLogRouter     TraceLogType = "router"
	TraceLogHTTPClient TraceLogType = "http_client"
)

var (
	// configFilePath app config file path (mutable, set by LoadConfig)
	configFilePath string
)
