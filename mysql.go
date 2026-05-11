package frame

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Dialect defines dialect for mysql
const Dialect = "mysql"

// DBMultiClient multi db conns
type DBMultiClient struct {
	clients map[string]*gorm.DB
}

// GetMySQLConn returns the mysql client map (backward-compatible function)
// Returns the database connection map of the active App.
func GetMySQLConn() *DBMultiClient {
	core := getActiveCore()
	if core == nil {
		return &DBMultiClient{clients: make(map[string]*gorm.DB)}
	}
	return core.legacyDBConns
}

// gormLogger implements GORM logger.Interface
type gormLogger struct {
	Log     *logrus.Logger
	Disable bool
}

func newGormLogger(config *Config) logger.Interface {
	return &gormLogger{Log: NewLogger(config), Disable: config.Mysql.DisableReqLog}
}

func (l *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.Disable = (level == logger.Silent)
	return &newLogger
}

func (l *gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.Disable {
		return
	}
	l.Log.WithFields(logrus.Fields{
		TraceIDKey: getTraceIDFromContext(ctx),
	}).Infof(msg, data...)
}

func (l *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.Disable {
		return
	}
	l.Log.WithFields(logrus.Fields{
		TraceIDKey: getTraceIDFromContext(ctx),
	}).Warnf(msg, data...)
}

func (l *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.Disable {
		return
	}
	l.Log.WithFields(logrus.Fields{
		TraceIDKey: getTraceIDFromContext(ctx),
	}).Errorf(msg, data...)
}

func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.Disable {
		return
	}
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			TraceIDKey: getTraceIDFromContext(ctx),
			"duration": time.Since(begin).Milliseconds(),
			"error":    err.Error(),
		}).Error(fc())
	} else {
		l.Log.WithFields(logrus.Fields{
			TraceIDKey: getTraceIDFromContext(ctx),
			"duration": time.Since(begin).Milliseconds(),
		}).Infoln(fc())
	}
}
