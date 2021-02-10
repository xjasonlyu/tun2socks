package engine

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/xjasonlyu/tun2socks/constant"
)

func showVersion() {
	fmt.Print(versionString())
	fmt.Print(releaseString())
}

func versionString() string {
	return fmt.Sprintf("%s %s\n", constant.Name, strings.TrimPrefix(constant.Version, "v"))
}

func releaseString() string {
	return fmt.Sprintf("%s/%s, %s, %s\n", runtime.GOOS, runtime.GOARCH, runtime.Version(), constant.BuildTime)
}
