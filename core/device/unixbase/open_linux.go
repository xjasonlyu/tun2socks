package unixbase

import (
	"fmt"

	"gvisor.dev/gvisor/pkg/tcpip/link/fdbased"

	"github.com/xjasonlyu/tun2socks/v2/core/device"
)

func open(u *Unix, offset int) (device.Device, error) {
	ep, err := fdbased.New(&fdbased.Options{
		FDs: []int{u.fd},
		MTU: u.mtu,
		// TUN only, ignore ethernet header.
		EthernetHeader: false,
	})
	if err != nil {
		return nil, fmt.Errorf("create endpoint: %w", err)
	}
	u.LinkEndpoint = ep

	return u, nil
}
