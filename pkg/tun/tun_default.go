// +build !darwin,!freebsd,!linux,!openbsd

package tun

import (
	"fmt"
	"runtime"
)

func CreateTUN(_ string, _ uint32) (Device, error) {
	return nil, fmt.Errorf("operation was not supported on %s", runtime.GOOS)
}
