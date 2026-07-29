//go:build unix

package fdbased

import (
	"fmt"
	"os"
	"strconv"

	"golang.org/x/sys/unix"
	"gvisor.dev/gvisor/pkg/tcpip/stack"

	"github.com/xjasonlyu/tun2socks/v2/core/device"
)

const defaultMTU = 1500

type FD struct {
	stack.LinkEndpoint

	fd     int
	file   *os.File // non-nil on darwin/bsd (iobased path); nil on linux
	mtu    uint32
	closed bool
}

func Open(name string, mtu uint32, offset int) (device.Device, error) {
	fd, err := strconv.Atoi(name)
	if err != nil {
		return nil, fmt.Errorf("cannot open fd: %s", name)
	}
	if mtu == 0 {
		mtu = defaultMTU
	}
	return open(fd, mtu, offset)
}

func (f *FD) Type() string {
	return Driver
}

func (f *FD) Name() string {
	return strconv.Itoa(f.fd)
}

func (f *FD) Close() {
	if !f.closed {
		defer f.LinkEndpoint.Close()
		// On darwin/bsd the iobased endpoint wraps the fd in an *os.File
		// and parks a goroutine in os.File.Read (runtime netpoller).
		// Closing via os.File.Close() triggers runtime_pollUnblock and
		// wakes that goroutine so dispatchLoop can exit and the link
		// endpoint's Wait() can return. A bare unix.Close(fd) does NOT
		// notify the netpoller, leaving the goroutine blocked forever.
		if f.file != nil {
			_ = f.file.Close()
		} else {
			_ = unix.Close(f.fd)
		}
		f.closed = true
	}
}

var _ device.Device = (*FD)(nil)
