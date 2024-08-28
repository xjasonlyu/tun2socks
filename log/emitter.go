package log

import (
	"io"
	"runtime"
	"strings"
	"time"

	glog "gvisor.dev/gvisor/pkg/log"
)

func init() {
	EnableStackLog(true)
}

func EnableStackLog(v bool) {
	if v {
		glog.SetTarget(&emitter{}) // built-in logger
	} else {
		glog.SetTarget(&glog.Writer{Next: io.Discard})
	}
}

type emitter struct{}

func (*emitter) level(level glog.Level) Level {
	return 1 - Level(level)
}

func (*emitter) prefix(format string) string {
	return "[STACK] " + format
}

func (e *emitter) Emit(depth int, level glog.Level, _ time.Time, format string, args ...any) {
	if _, file, line, ok := runtime.Caller(depth + 1); ok {
		// Ignore: gvisor.dev/gvisor/pkg/tcpip/adapters/gonet/gonet.go:457
		if line == 457 && strings.HasSuffix(file, "gonet/gonet.go") {
			return
		}
	}
	logf(e.level(level), e.prefix(format), args...)
}
