//go:build unix && !linux

package fdbased

import (
	"fmt"
	"os"

	"github.com/xjasonlyu/tun2socks/v2/core/device"
	"github.com/xjasonlyu/tun2socks/v2/core/device/iobased"
)

func open(fd int, mtu uint32, offset int) (device.Device, error) {
	f := &FD{fd: fd, mtu: mtu}
	file := os.NewFile(uintptr(fd), f.Name())
	ep, err := iobased.New(file, mtu, offset)
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("create endpoint: %w", err)
	}
	f.LinkEndpoint = ep
	// Keep the *os.File reference so FD.Close() can close it via the
	// runtime netpoller path and unblock the iobased dispatchLoop.
	f.file = file

	return f, nil
}
