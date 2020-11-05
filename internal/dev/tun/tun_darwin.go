package tun

import (
	"fmt"
	"io"
	"syscall"
	"unsafe"

	"github.com/songgao/water"
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
	// open datagram socket
	fd, err := syscall.Socket(
		syscall.AF_INET,
		syscall.SOCK_DGRAM,
		0,
	)
	if err != nil {
		return 0, err
	}

	defer syscall.Close(fd)

	// do ioctl call
	var ifr struct {
		name [16]byte
		mtu  uint32
	}
	copy(ifr.name[:], name)

	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(syscall.SIOCGIFMTU),
		uintptr(unsafe.Pointer(&ifr)),
	)
	if errno != 0 {
		return 0, fmt.Errorf("get MTU on %s: %s", name, errno.Error())
	}

	return ifr.mtu, nil
}
