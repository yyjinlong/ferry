// copyright @ 2020 ops inc.
//
// author: jinlong yang
//

package log

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Fields map[string]interface{}

var (
	logging *logrus.Logger = logrus.New()
	logger  *logrus.Entry
)

func InitLogger(logFile string) {
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logrus.Fatalf("Open log file failed: %s", err)
	}

	logging.SetFormatter(&logrus.JSONFormatter{})
	if gin.Mode() == gin.DebugMode {
		logging.SetOutput(os.Stdout)
	} else {
		logging.SetOutput(f)
	}
	logging.SetLevel(logrus.InfoLevel)

	// NOTE: 初始默认设置logger实例
	logger = logging.WithFields(logrus.Fields{})
}

func InitFields(fields Fields) {
	fieldInfo := logrus.Fields{}
	for k, v := range fields {
		fieldInfo[k] = v
	}
	logger = logging.WithFields(fieldInfo)
}

func Debug(args ...interface{}) {
	logger.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

func Info(args ...interface{}) {
	logger.Info(args...)
}

func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

func Warn(args ...interface{}) {
	logger.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

func Error(args ...interface{}) {
	logger.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

func Panic(args ...interface{}) {
	logger.Panic(args...)
}

func Panicf(format string, args ...interface{}) {
	logger.Panicf(format, args...)
}

func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
}
