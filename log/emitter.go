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

func (emitter) Emit(depth int, level glog.Level, _ time.Time, format string, args ...any) {
	if _, file, line, ok := runtime.Caller(depth + 1); ok {
		// Ignore (*gonet.TCPConn).RemoteAddr() warning: `ep.GetRemoteAddress() failed`.
		if line == 457 && strings.HasSuffix(file, "/pkg/tcpip/adapters/gonet/gonet.go") {
			return
		}
	}
	logf(Level(level)+2, "[STACK] "+format, args...)
}
