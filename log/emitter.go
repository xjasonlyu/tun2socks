package log

import (
	"time"

	glog "gvisor.dev/gvisor/pkg/log"
)

var _globalE = &emitter{}

func init() {
	glog.SetTarget(_globalE)
}

type emitter struct {
	logger *SugaredLogger
}

func (e *emitter) setLogger(logger *SugaredLogger) {
	e.logger = logger.WithOptions(pkgCallerSkip)
	glog.SetLevel(1 - glog.Level(e.logger.Level()))
}

func (e *emitter) logf(level glog.Level, format string, args ...any) {
	e.logger.Logf(1-Level(level), "[STACK] "+format, args...)
}

func (e *emitter) Emit(_ int, level glog.Level, _ time.Time, format string, args ...any) {
	e.logf(level, format, args...)
}
