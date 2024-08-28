package log

import (
	"runtime"
	"strings"
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
}

func (e *emitter) logf(level glog.Level, format string, args ...any) {
	e.logger.Logf(1-Level(level), "[STACK] "+format, args...)
}

func (e *emitter) Emit(depth int, level glog.Level, _ time.Time, format string, args ...any) {
	if _, file, line, ok := runtime.Caller(depth + 1); ok {
		// Ignore: gvisor.dev/gvisor/pkg/tcpip/adapters/gonet/gonet.go:457
		if line == 457 && strings.HasSuffix(file, "gonet/gonet.go") {
			return
		}
	}
	e.logf(level, format, args...)
}
