package dialer

import (
	"errors"
	"syscall"
)

var _ SocketOption = SocketOptionFunc(nil)

type SocketOption interface {
	Apply(network, address string, c syscall.RawConn) error
}

// SocketOptionFunc adapts a function to a SocketOption.
type SocketOptionFunc func(network, address string, c syscall.RawConn) error

func (f SocketOptionFunc) Apply(network, address string, c syscall.RawConn) error {
	return f(network, address, c)
}

// UnsupportedSocketOption is a sentinel SocketOption that always reports
// ErrUnsupported when applied.
var UnsupportedSocketOption = SocketOptionFunc(unsupportedSocketOpt)

func unsupportedSocketOpt(_, _ string, _ syscall.RawConn) error {
	return errors.ErrUnsupported
}

// rawConnControl runs f with the file descriptor obtained via RawConn.Control
// and correctly propagates errors returned from f.
func rawConnControl(c syscall.RawConn, f func(uintptr) error) error {
	var innerErr error
	if err := c.Control(func(fd uintptr) {
		innerErr = f(fd)
	}); err != nil {
		return err
	}
	return innerErr
}
