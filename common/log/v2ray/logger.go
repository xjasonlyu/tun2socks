package v2ray

import (
	"fmt"
	"os"

	vlog "v2ray.com/core/common/log"

	"github.com/xjasonlyu/tun2socks/common/log"
)

type v2rayLogger struct {
	handler vlog.Handler
	level   vlog.Severity
}

func NewV2RayLogger() log.Logger {
	handler := vlog.NewLogger(vlog.CreateStdoutLogWriter())
	return &v2rayLogger{handler: handler, level: vlog.Severity_Info}
}

func (l *v2rayLogger) SetLevel(level log.LogLevel) {
	switch level {
	case log.DEBUG:
		l.level = vlog.Severity_Debug
	case log.INFO:
		l.level = vlog.Severity_Info
	case log.WARN:
		l.level = vlog.Severity_Warning
	case log.ERROR:
		l.level = vlog.Severity_Error
	case log.NONE:
		l.level = vlog.Severity_Unknown
	}
}

func (l *v2rayLogger) Debugf(msg string, args ...interface{}) {
	l.output(vlog.Severity_Debug, msg, args...)
}

func (l *v2rayLogger) Infof(msg string, args ...interface{}) {
	l.output(vlog.Severity_Info, msg, args...)
}

func (l *v2rayLogger) Warnf(msg string, args ...interface{}) {
	l.output(vlog.Severity_Warning, msg, args...)
}

func (l *v2rayLogger) Errorf(msg string, args ...interface{}) {
	l.output(vlog.Severity_Error, msg, args...)
}

func (l *v2rayLogger) Fatalf(msg string, args ...interface{}) {
	l.output(vlog.Severity_Unknown, msg, args...)
	os.Exit(1)
}

func (l *v2rayLogger) output(level vlog.Severity, msg string, args ...interface{}) {
	if level <= l.level {
		l.handler.Handle(&vlog.GeneralMessage{
			Severity: level,
			Content:  fmt.Sprintf(msg, args...),
		})
	}
}

func init() {
	log.RegisterLogger(NewV2RayLogger())
}
