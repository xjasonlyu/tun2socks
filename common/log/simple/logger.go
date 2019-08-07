package simple

import (
	goLog "log"

	"github.com/xjasonlyu/tun2socks/common/log"
)

type simpleLogger struct {
	level log.Level
}

func NewSimpleLogger() log.Logger {
	return &simpleLogger{log.INFO}
}

func (l *simpleLogger) SetLevel(level log.Level) {
	l.level = level
}

func (l *simpleLogger) Debugf(msg string, args ...interface{}) {
	if l.level <= log.DEBUG {
		l.output(msg, args...)
	}
}

func (l *simpleLogger) Infof(msg string, args ...interface{}) {
	if l.level <= log.INFO {
		l.output(msg, args...)
	}
}

func (l *simpleLogger) Warnf(msg string, args ...interface{}) {
	if l.level <= log.WARN {
		l.output(msg, args...)
	}
}

func (l *simpleLogger) Errorf(msg string, args ...interface{}) {
	if l.level <= log.ERROR {
		l.output(msg, args...)
	}
}

func (l *simpleLogger) Fatalf(msg string, args ...interface{}) {
	goLog.Fatalf(msg, args...)
}

func (l *simpleLogger) output(msg string, args ...interface{}) {
	goLog.Printf(msg, args...)
}

func init() {
	log.RegisterLogger(NewSimpleLogger())
}
