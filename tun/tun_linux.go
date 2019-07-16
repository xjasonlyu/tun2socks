package tun

import (
	"io"

	"github.com/songgao/water"
)

func OpenTunDevice(name, addr, gw, mask string, dnsServers []string) (io.ReadWriteCloser, error) {
	cfg := water.Config{
		DeviceType: water.TUN,
	}
	cfg.Name = name
	tunDev, err := water.New(cfg)
	if err != nil {
		return nil, err
	}
	name = tunDev.Name()
	return tunDev, nil
}
