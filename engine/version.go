package engine

import (
	"fmt"
	"runtime"
	"strings"

	V "github.com/xjasonlyu/tun2socks/constant"
)

func showVersion() {
	fmt.Print(versionString())
	fmt.Print(releaseString())
}

func versionString() string {
	return fmt.Sprintf("%s-%s\n", V.Name, strings.TrimPrefix(V.Version, "v"))
}

func releaseString() string {
	return fmt.Sprintf("%s/%s, %s, %s\n", runtime.GOOS, runtime.GOARCH, runtime.Version(), V.GitCommit)
}
