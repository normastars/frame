package frame

import (
	"github.com/sirupsen/logrus"
)

var defaultJSONLogFormatter = &logrus.JSONFormatter{
	// CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
	// 	return frame.Function, path.Base(frame.File)
	// },
}

// init config project logrus log
func init() {
	// default log config
	logrus.SetReportCaller(true)
	// set file name
	logrus.SetFormatter(defaultJSONLogFormatter)
}
