package tun

import (
	"fmt"
	"io"

	"github.com/songgao/water"
	"golang.org/x/sys/unix"
	"gvisor.dev/gvisor/pkg/tcpip/stack"

	"github.com/xjasonlyu/tun2socks/pkg/link/rwc"
)

func Open(name string) (ep stack.LinkEndpoint, c io.Closer, err error) {
	config := water.Config{
		DeviceType: water.TUN,
	}
	config.Name = name

	var ifce *water.Interface
	ifce, err = water.New(config)
	if err != nil {
		return
	}

	var mtu uint32
	mtu, err = getMTU(name)
	if err != nil {
		return
	}

	ep, err = rwc.New(ifce, mtu)
	c = ifce

	return
}

func getMTU(name string) (uint32, error) {
	fd, err := unix.Socket(
		unix.AF_INET,
		unix.SOCK_DGRAM,
		0,
	)

	if err != nil {
		return 0, err
	}

	defer unix.Close(fd)

	ifr, err := unix.IoctlGetIfreqMTU(fd, name)
	if err != nil {
		return 0, fmt.Errorf("get MTU on %s: %w", name, err)
	}

	return uint32(ifr.MTU), nil
}
