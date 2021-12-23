//go:build !linux

package tun

import (
	"fmt"

	"github.com/xjasonlyu/tun2socks/v2/core/device"
	"github.com/xjasonlyu/tun2socks/v2/core/device/rwbased"

	"golang.zx2c4.com/wireguard/tun"
)

type TUN struct {
	*rwbased.Endpoint

	nt   *tun.NativeTun
	mtu  uint32
	name string
}

func Open(name string, mtu uint32) (device.Device, error) {
	t := &TUN{name: name, mtu: mtu}

	forcedMTU := defaultMTU
	if t.mtu > 0 {
		forcedMTU = int(t.mtu)
	}

	nt, err := tun.CreateTUN(t.name, forcedMTU)
	if err != nil {
		return nil, fmt.Errorf("create tun: %w", err)
	}
	t.nt = nt.(*tun.NativeTun)

	_mtu, err := nt.MTU()
	if err != nil {
		return nil, fmt.Errorf("get mtu: %w", err)
	}
	t.mtu = uint32(_mtu)

	ep, err := rwbased.New(t, t.mtu)
	if err != nil {
		return nil, fmt.Errorf("create endpoint: %w", err)
	}
	t.Endpoint = ep

	return t, nil
}

func (t *TUN) Name() string {
	name, _ := t.nt.Name()
	return name
}

func (t *TUN) Close() error {
	return t.nt.Close()
}
