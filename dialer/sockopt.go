package dialer

import (
	"syscall"
)

var _ SocketOption = SocketOptionFunc(nil)

type SocketOption interface {
	Apply(network, address string, c syscall.RawConn) error
}

type SocketOptionFunc func(network, address string, c syscall.RawConn) error

func (f SocketOptionFunc) Apply(network, address string, c syscall.RawConn) error {
	return f(network, address, c)
}

var NopSocketOption = SocketOptionFunc(func(_, _ string, _ syscall.RawConn) error { return nil })

func control(c syscall.RawConn, f func(uintptr) error) error {
	var innerErr error
	err := c.Control(func(fd uintptr) {
		innerErr = f(fd)
	})
	if innerErr != nil {
		err = innerErr
	}
	return err
}
