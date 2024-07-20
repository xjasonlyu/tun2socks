//go:build !linux

package tun

import (
	"fmt"
	"sync"

	"golang.zx2c4.com/wireguard/tun"

	"github.com/xjasonlyu/tun2socks/v2/core/device"
	"github.com/xjasonlyu/tun2socks/v2/core/device/iobased"
)

type TUN struct {
	*iobased.Endpoint

	nt     *tun.NativeTun
	mtu    uint32
	name   string
	offset int

	rSizes []int
	rBuffs [][]byte
	wBuffs [][]byte
	rMutex sync.Mutex
	wMutex sync.Mutex
}

func Open(name string, mtu uint32) (_ device.Device, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("open tun: %v", r)
		}
	}()

	t := &TUN{
		name:   name,
		mtu:    mtu,
		offset: offset,
		rSizes: make([]int, 1),
		rBuffs: make([][]byte, 1),
		wBuffs: make([][]byte, 1),
	}

	forcedMTU := defaultMTU
	if t.mtu > 0 {
		forcedMTU = int(t.mtu)
	}

	nt, err := createTUN(t.name, forcedMTU)
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
	t.rMutex.Lock()
	defer t.rMutex.Unlock()
	t.rBuffs[0] = packet
	_, err := t.nt.Read(t.rBuffs, t.rSizes, t.offset)
	return t.rSizes[0], err
}

func (t *TUN) Write(packet []byte) (int, error) {
	t.wMutex.Lock()
	defer t.wMutex.Unlock()
	t.wBuffs[0] = packet
	return t.nt.Write(t.wBuffs, t.offset)
}

func (t *TUN) Name() string {
	name, _ := t.nt.Name()
	return name
}

func (t *TUN) Close() {
	defer t.Endpoint.Close()
	_ = t.nt.Close()
}
