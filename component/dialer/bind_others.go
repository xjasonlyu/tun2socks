//go:build !linux && !darwin

package dialer

import (
	"errors"
	"net"
	"syscall"
)

func bindToInterface(_ *net.Interface) controlFunc {
	return func(string, string, syscall.RawConn) error {
		return errors.New("unsupported platform")
	}
}
