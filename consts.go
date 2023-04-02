package frame

import (
	"github.com/sirupsen/logrus"
	"gorm.io/gorm/logger"
)

// frame consts list
const (
	TraceIDKey  string = "trace_id"
	ModeJSON           = "json"
	ModelText          = "text"
	TraceLogKey        = "req_msg"
)

var (
	logLevelTrace = "trace"
	logLevelDebug = "debug"
	logLevelInfo  = "info"
	logLevelWarn  = "warn"
	logLevelError = "error"
	logLevelFatal = "fatal"
	logLevelPanic = "panic"
)

var (
	defaultMetricName  = "metric"
	defaultMetricsName = "metrics"
	defaultMetricPath  = "/metrics"
	defaultMetricPort  = ":9090"
	defaultMetricPort2 = "9090"
	defaultBusName     = "server"
	defaultBusPort     = "8080"
	defaultBusPort2    = ":8080"
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

var (
	codeKey          = "code"
	msgKey           = "message"
	successMsg       = "ok"
	successCode      = "0"
	defaultEmptyPage = PageResults{
		Results: make([]interface{}, 0),
	}
)

// config type
var (
	configTypeYaml    = "yaml"
	configTypeYal     = "yml"
	configTypeJSON    = "json" // default json
	configPath        = "CONFPATH"
	configType        = "CONFTYPE" // default ./conf/default.json
	configDefaultPath = "./conf/default.json"
)

var defaultMetricHTTPConfig = HTTPServerConfig{
	Name: defaultMetricName,
	Port: defaultMetricPort2,
}

var defaultBusHTTPConfig = HTTPServerConfig{
	Name: defaultBusName,
	Port: defaultMetricPort2,
}

// TraceLogType trace lo type
type TraceLogType string

// trace type
var (
	TraceLogRouter     TraceLogType = "router"
	TraceLogHTTPClient TraceLogType = "http_client"
)
