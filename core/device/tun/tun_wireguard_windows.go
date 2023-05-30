package tun

import (
	"golang.zx2c4.com/wireguard/tun"
)

const (
	offset     = 0
	defaultMTU = 0 /* auto */
)

func createTUN(name string, mtu int) (tun.Device, error) {
	return tun.CreateTUN(name, mtu)
}
