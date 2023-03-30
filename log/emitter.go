package log

import (
	"io"
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

func (emitter) Emit(_ int, level glog.Level, _ time.Time, format string, args ...any) {
	logf(Level(level)+2, "[STACK] "+format, args...)
}
