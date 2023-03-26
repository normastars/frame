package frame

import (
	"github.com/sirupsen/logrus"
)

var defaultLogFormatter = &logrus.JSONFormatter{
	// CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
	// 	return frame.Function, path.Base(frame.File)
	// },
}

func init() {
	// default log config
	logrus.SetReportCaller(true)
	// set file name
	logrus.SetFormatter(defaultLogFormatter)
}
