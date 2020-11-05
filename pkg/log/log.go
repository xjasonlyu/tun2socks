package log

import (
	"os"
	"sync/atomic"

	"github.com/sirupsen/logrus"
)

var (
	// defaultLevel is package default loglevel.
	defaultLevel = InfoLevel
)

func init() {
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)
}

func SetLevel(level Level) {
	atomic.StoreUint32((*uint32)(&defaultLevel), uint32(level))
}

func Debugf(format string, args ...interface{}) {
	logf(DebugLevel, format, args...)
}

func Infof(format string, args ...interface{}) {
	logf(InfoLevel, format, args...)
}

func Warnf(format string, args ...interface{}) {
	logf(WarnLevel, format, args...)
}

func Errorf(format string, args ...interface{}) {
	logf(ErrorLevel, format, args...)
}

func Fatalf(format string, args ...interface{}) {
	logrus.Fatalf(format, args...)
}

func logf(level Level, format string, args ...interface{}) {
	event := newEvent(level, format, args...)
	if event.Level > defaultLevel {
		return
	}

	switch level {
	case DebugLevel:
		logrus.WithTime(event.Time).Debugln(event.Message)
	case InfoLevel:
		logrus.WithTime(event.Time).Infoln(event.Message)
	case WarnLevel:
		logrus.WithTime(event.Time).Warnln(event.Message)
	case ErrorLevel:
		logrus.WithTime(event.Time).Errorln(event.Message)
	}
}
