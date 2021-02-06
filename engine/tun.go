// +build !windows

package engine

import (
	"net/url"

	"github.com/xjasonlyu/tun2socks/device"
	"github.com/xjasonlyu/tun2socks/device/tun"
)

func openTUN(u *url.URL, mtu uint32) (device.Device, error) {
	name := u.Host
	return tun.Open(tun.WithName(name), tun.WithMTU(mtu))
}
