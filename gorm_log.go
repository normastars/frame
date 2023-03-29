package frame

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm/logger"
)

type gormLogger struct {
	Log     *logrus.Logger
	TraceID string
}

func newGormLogger(traceID string) logger.Interface {
	return &gormLogger{Log: logrus.New(), TraceID: traceID}
}

func (l *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

func (l *gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.Log.WithFields(logrus.Fields{
		TraceID: getTraceIDFromContext(ctx),
	}).Infof(msg, data...)
}

func (l *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.Log.WithFields(logrus.Fields{
		TraceID: getTraceIDFromContext(ctx),
	}).Warnf(msg, data...)
}

func (l *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.Log.WithFields(logrus.Fields{
		TraceID: getTraceIDFromContext(ctx),
	}).Errorf(msg, data...)
}

func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			TraceID:    getTraceIDFromContext(ctx),
			"duration": time.Since(begin).String(),
			"error":    err.Error(),
		}).Error(fc())
	} else {
		l.Log.WithFields(logrus.Fields{
			TraceID:    getTraceIDFromContext(ctx),
			"duration": time.Since(begin).String(),
		}).Infoln(fc())
	}
}

func getTraceIDFromContext(ctx context.Context) string {
	traceID := ctx.Value(TraceID)
	return traceID.(string)
}
