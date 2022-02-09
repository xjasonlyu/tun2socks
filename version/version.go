package version

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/xjasonlyu/tun2socks/v2/internal/debug"
)

const Name = "tun2socks"

var (
	// Version can be set at link time by executing
	// the command: `git describe --abbrev=0 --tags HEAD`
	Version string

	// GitCommit can be set at link time by executing
	// the command: `git rev-parse --short HEAD`
	GitCommit string
)

func versionize(s string) string {
	return strings.TrimPrefix(s, "v")
}

func String() string {
	if !debug.Debug() {
		return fmt.Sprintf("%s-%s", Name, versionize(Version))
	}
	return fmt.Sprintf("%s-%s (debug)", Name, versionize(Version))
}

func BuildString() string {
	return fmt.Sprintf("%s/%s, %s, %s", runtime.GOOS, runtime.GOARCH, runtime.Version(), GitCommit)
}
