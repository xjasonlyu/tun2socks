// +build !darwin,!linux

package tun

import (
	"fmt"
	"io"
	"runtime"

	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

func Open(_ string) (stack.LinkEndpoint, io.Closer, error) {
	return nil, nil, fmt.Errorf("operation was not supported on %s", runtime.GOOS)
}
