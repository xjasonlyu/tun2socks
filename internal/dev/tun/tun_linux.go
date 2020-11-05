package tun

import (
	"io"
	"syscall"

	"gvisor.dev/gvisor/pkg/tcpip/link/fdbased"
	"gvisor.dev/gvisor/pkg/tcpip/link/rawfile"
	"gvisor.dev/gvisor/pkg/tcpip/link/tun"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

type closeFunc func() error

func (f closeFunc) Close() error {
	return f()
}

func Open(name string) (ep stack.LinkEndpoint, c io.Closer, err error) {
	var fd int
	fd, err = tun.Open(name)
	if err != nil {
		return
	}

	var mtu uint32
	mtu, err = rawfile.GetMTU(name)
	if err != nil {
		return
	}

	ep, err = fdbased.New(&fdbased.Options{
		FDs:            []int{fd},
		MTU:            mtu,
		EthernetHeader: false,
	})

	c = closeFunc(func() error {
		return syscall.Close(fd)
	})

	return
}
