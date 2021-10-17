package fd

import (
	"strconv"

	"github.com/xjasonlyu/tun2socks/core/device"
	"golang.org/x/sys/unix"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

const Driver = "fd"

type FD struct {
	stack.LinkEndpoint

	fd  int
	mtu uint32
}

func (f *FD) Type() string {
	return Driver
}

func (f *FD) Name() string {
	return strconv.Itoa(f.fd)
}

func (f *FD) Close() error {
	return unix.Close(f.fd)
}

var _ device.Device = (*FD)(nil)
