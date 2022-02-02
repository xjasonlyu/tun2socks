package version

import (
	"fmt"
	"runtime"
	"strings"
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
	return fmt.Sprintf("%s-%s", Name, versionize(Version))
}

func BuildString() string {
	return fmt.Sprintf("%s/%s, %s, %s", runtime.GOOS, runtime.GOARCH, runtime.Version(), GitCommit)
}
