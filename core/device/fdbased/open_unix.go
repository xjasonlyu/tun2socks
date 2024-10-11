//go:build unix && !linux

package fdbased

import (
	"fmt"
	"os"
	"runtime"

	"github.com/xjasonlyu/tun2socks/v2/core/device"
	"github.com/xjasonlyu/tun2socks/v2/core/device/iobased"
)

func open(fd int, mtu uint32, offset int) (device.Device, error) {
	f := &FD{fd: fd, mtu: mtu}
	// fd offset in ios
	// https://stackoverflow.com/questions/69260852/ios-network-extension-packet-parsing/69487795#69487795
	if offset == 0 && runtime.GOOS == "ios" {
		offset = 4
	}
	ep, err := iobased.New(os.NewFile(uintptr(fd), f.Name()), mtu, offset)
	if err != nil {
		return nil, fmt.Errorf("create endpoint: %w", err)
	}
	f.LinkEndpoint = ep

	return f, nil
}
