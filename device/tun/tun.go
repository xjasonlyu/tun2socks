// Package tun provides TUN which implemented device.Device interface.
package tun

import (
	"github.com/xjasonlyu/tun2socks/device"
)

const driverType = "tun"

func (t *TUN) Type() string {
	return driverType
}

var _ device.Device = (*TUN)(nil)
