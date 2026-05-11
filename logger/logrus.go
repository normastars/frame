// Package logger provides Logrus logger wrappers (instance-level).
// Eliminates the global reqLogger; each instance holds its own Logger.
package logger

import (
	"path"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm/logger"
)

var defaultJSONLogFormatter = &logrus.JSONFormatter{
	CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
		return frame.Function, path.Base(frame.File)
	},
}

// Manager manages logrus loggers (instance-level)
type Manager struct {
	Logger   *logrus.Logger
	logLevel string
	logMode  string
}

// NewManager creates a new log manager
func NewManager(logLevel, logMode string) *Manager {
	l := logrus.New()
	l.SetReportCaller(true)
	logLevel = strings.ToLower(logLevel)
	logMode = strings.ToLower(logMode)

	if logMode == "" || logMode == "json" {
		l.SetFormatter(defaultJSONLogFormatter)
	}
	if level, ok := logLevelMap[logLevel]; ok {
		l.SetLevel(level)
	} else {
		l.SetLevel(logrus.InfoLevel)
	}

	return &Manager{
		Logger:   l,
		logLevel: logLevel,
		logMode:  logMode,
	}
}

// NewLogger creates a new Logrus Logger instance from config
func (m *Manager) NewLogger() *logrus.Logger {
	l := logrus.New()
	l.SetReportCaller(true)
	if m.logMode == "" || m.logMode == "json" {
		l.SetFormatter(defaultJSONLogFormatter)
	}
	if level, ok := logLevelMap[m.logLevel]; ok {
		l.SetLevel(level)
	} else {
		l.SetLevel(logrus.InfoLevel)
	}
	return l
}

// Log2GormLevel converts logrus log level to GORM log level
func Log2GormLevel(level string) logger.LogLevel {
	level = strings.ToLower(level)
	le, ok := gormLogLevelMap[level]
	if ok {
		return le
	}
	return logger.Silent
}

var logLevelMap = map[string]logrus.Level{
	"panic": logrus.PanicLevel,
	"fatal": logrus.FatalLevel,
	"error": logrus.ErrorLevel,
	"warn":  logrus.WarnLevel,
	"info":  logrus.InfoLevel,
	"debug": logrus.DebugLevel,
	"trace": logrus.TraceLevel,
}

var gormLogLevelMap = map[string]logger.LogLevel{
	"error": logger.Error,
	"warn":  logger.Warn,
	"info":  logger.Info,
}
