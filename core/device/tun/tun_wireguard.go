//go:build !linux

package tun

import (
	"fmt"

	"github.com/xjasonlyu/tun2socks/v2/core/device"
	"github.com/xjasonlyu/tun2socks/v2/core/device/iobased"

	"golang.zx2c4.com/wireguard/tun"
)

type TUN struct {
	*iobased.Endpoint

	nt     *tun.NativeTun
	mtu    uint32
	name   string
	offset int
}

func Open(name string, mtu uint32) (_ device.Device, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("open tun: %v", r)
		}
	}()

	t := &TUN{name: name, mtu: mtu, offset: offset}

	forcedMTU := defaultMTU
	if t.mtu > 0 {
		forcedMTU = int(t.mtu)
	}

	nt, err := tun.CreateTUN(t.name, forcedMTU)
	if err != nil {
		return nil, fmt.Errorf("create tun: %w", err)
	}
	t.nt = nt.(*tun.NativeTun)

	tunMTU, err := nt.MTU()
	if err != nil {
		return nil, fmt.Errorf("get mtu: %w", err)
	}
	t.mtu = uint32(tunMTU)

	ep, err := iobased.New(t, t.mtu, offset)
	if err != nil {
		return nil, fmt.Errorf("create endpoint: %w", err)
	}
	t.Endpoint = ep

	return t, nil
}

func (t *TUN) Read(packet []byte) (int, error) {
	return t.nt.Read(packet, t.offset)
}

func (t *TUN) Write(packet []byte) (int, error) {
	return t.nt.Write(packet, t.offset)
}

func (t *TUN) Name() string {
	name, _ := t.nt.Name()
	return name
}

func (t *TUN) Close() error {
	defer t.Endpoint.Close()
	return t.nt.Close()
}
