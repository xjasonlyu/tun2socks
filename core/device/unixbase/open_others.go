//go:build !(linux && amd64) && !(linux && arm64) && !windows

package unixbase

import (
	"fmt"
	"github.com/xjasonlyu/tun2socks/v2/core/device"
	"github.com/xjasonlyu/tun2socks/v2/core/device/iobased"
	"os"
)

func open(u *Unix, offset int) (device.Device, error) {

	ep, err := iobased.New(os.NewFile(uintptr(u.fd), u.Fd()), u.mtu, offset)
	if err != nil {
		return nil, fmt.Errorf("create endpoint: %w", err)
	}
	u.LinkEndpoint = ep

	return u, nil
}
