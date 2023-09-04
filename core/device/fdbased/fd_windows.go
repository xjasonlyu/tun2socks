package fdbased

import (
	"errors"

	"github.com/xjasonlyu/tun2socks/v2/core/device"
)

func Open(name string, mtu uint32, offset int) (device.Device, error) {
	return nil, errors.ErrUnsupported
}
