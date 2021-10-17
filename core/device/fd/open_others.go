//go:build !linux

package fd

import (
	"fmt"
	"os"

	"github.com/xjasonlyu/tun2socks/core/device"
	"github.com/xjasonlyu/tun2socks/core/device/rwbased"
)

func Open(fd int, mtu uint32) (device.Device, error) {
	f := &FD{fd: fd, mtu: mtu}

	ep, err := rwbased.New(os.NewFile(uintptr(fd), f.Name()), mtu)
	if err != nil {
		return nil, fmt.Errorf("create endpoint: %w", err)
	}
	f.LinkEndpoint = ep

	return f, nil
}
