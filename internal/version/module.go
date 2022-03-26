package version

import (
	"runtime/debug"
)

// Info returns project dependencies as []*debug.Module.
func Info() []*debug.Module {
	bi, _ := debug.ReadBuildInfo()
	return bi.Deps
}
