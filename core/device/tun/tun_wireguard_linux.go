//go:build linux && !(amd64 || arm64)

package tun

import (
	"fmt"
	"os"

	"golang.zx2c4.com/wireguard/tun"
	gun "gvisor.dev/gvisor/pkg/tcpip/link/tun"
)

const (
	offset     = 0 /* IFF_NO_PI */
	defaultMTU = 1500
)

func createTUN(name string, mtu int) (tun.Device, error) {
	nfd, err := gun.Open(name)
	if err != nil {
		return nil, fmt.Errorf("create tun: %w", err)
	}

	fd := os.NewFile(uintptr(nfd), "/dev/net/tun")
	return tun.CreateTUNFromFile(fd, mtu)
}
