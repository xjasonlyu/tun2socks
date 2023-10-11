//go:build !windows

package unixbase

import (
	"errors"
	"fmt"
	"net"
	"strconv"

	"golang.org/x/sys/unix"
	"gvisor.dev/gvisor/pkg/tcpip/stack"

	"github.com/xjasonlyu/tun2socks/v2/core/device"
)

type Unix struct {
	stack.LinkEndpoint
	path string
	fd   int
	mtu  uint32
}

/*
for example:
device: unix:///tmp/unix_socket
*/

func unix2Fd(path string) (int, error) {
	conn, err := net.Dial("unix", path)
	if err != nil {
		return 0, fmt.Errorf("unix socket [%s] connection error: %v", path, conn)
	}

	unixConn, ok := conn.(*net.UnixConn)
	if !ok {
		return 0, errors.New("unexpected")
	}

	fd, err := unixConn.File()
	if err != nil {
		return 0, errors.New("unexpected")
	}

	return int(fd.Fd()), nil
}

const defaultMTU = 1500

func Open(path string, mtu uint32, offset int) (device.Device, error) {
	fd, err := unix2Fd(path)
	if err != nil {
		return nil, err
	}

	if mtu == 0 {
		mtu = defaultMTU
	}

	f := &Unix{
		fd:   fd,
		path: path,
		mtu:  mtu,
	}

	return open(f, offset)
}

func (f *Unix) Type() string {
	return Driver
}

func (f *Unix) Name() string {
	return f.path
}

func (f *Unix) Fd() string {
	return strconv.Itoa(f.fd)
}

func (f *Unix) Close() error {
	return unix.Close(f.fd)
}

var _ device.Device = (*Unix)(nil)
