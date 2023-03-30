package frame

import (
	"github.com/sirupsen/logrus"
	"gorm.io/gorm/logger"
)

// frame consts list
const (
	TraceIDKey string = "trace_id"
	ModeJSON          = "json"
	ModelText         = "text"
)

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

var (
	codeKey          = "code"
	msgKey           = "message"
	successMsg       = "ok"
	successCode      = "0"
	defaultEmptyPage = PageResults{
		Results: make([]interface{}, 0),
	}
)
