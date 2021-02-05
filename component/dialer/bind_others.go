// +build !linux,!darwin

package dialer

import (
	"errors"
	"syscall"
)

func bindToInterface(network, address string, c syscall.RawConn) error {
	return errors.New("unsupported platform")
}
