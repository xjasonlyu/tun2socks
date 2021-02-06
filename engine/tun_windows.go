package engine

import (
	"net/url"

	"github.com/xjasonlyu/tun2socks/device"
	"github.com/xjasonlyu/tun2socks/device/tun"
)

func openTUN(u *url.URL, mtu uint32) (device.Device, error) {
	/*
	  e.g. tun://TUN0/?id=tap0901&network=10.10.10.10/24
	*/

	name := u.Host

	componentID := u.Query().Get("id")
	network := u.Query().Get("network")

	if componentID == "" {
		componentID = "tap0901" /* default */
	}
	if network == "" {
		network = "10.10.10.10/24" /* default */
	}

	return tun.Open(
		tun.WithName(name),
		tun.WithMTU(mtu),
		tun.WithComponentID(componentID),
		tun.WithNetwork(network),
	)
}
