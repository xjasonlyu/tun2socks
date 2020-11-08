// +build darwin freebsd openbsd

package tun

import (
	"golang.zx2c4.com/wireguard/tun"

	"github.com/xjasonlyu/clash/common/pool"
	"github.com/xjasonlyu/tun2socks/pkg/link/rwc"
)

const offset = 4

type unixTun struct {
	*rwc.Endpoint

	device tun.Device
}

func CreateTUN(name string, n uint32) (Device, error) {
	device, err := tun.CreateTUN(name, int(n))
	if err != nil {
		return nil, err
	}

	mtu, err := device.MTU()
	if err != nil {
		return nil, err
	}

	ut := &unixTun{
		device: device,
	}

	if ut.Endpoint, err = rwc.New(ut, uint32(mtu)); err != nil {
		return nil, err
	}

	return ut, nil
}

func (t *unixTun) Read(packet []byte) (n int, err error) {
	buf := pool.Get(offset + len(packet))
	defer pool.Put(buf)

	if n, err = t.device.Read(buf, offset); err != nil {
		return
	}

	copy(packet, buf[offset:offset+n])
	return
}

func (t *unixTun) Write(packet []byte) (int, error) {
	buf := pool.Get(offset + len(packet))
	defer pool.Put(buf)

	copy(buf[offset:], packet)
	return t.device.Write(buf[:offset+len(packet)], offset)
}

func (t *unixTun) Name() string {
	name, _ := t.device.Name()
	return name
}

func (t *unixTun) Close() error {
	return t.device.Close()
}
