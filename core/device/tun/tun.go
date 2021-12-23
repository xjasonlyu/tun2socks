// Package tun provides TUN which implemented device.Device interface.
package tun

import (
	"github.com/xjasonlyu/tun2socks/v2/core/device"
)

const Driver = "tun"

func (t *TUN) Type() string {
	return Driver
}

var _ device.Device = (*TUN)(nil)
