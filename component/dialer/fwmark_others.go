//go:build !linux

package dialer

import (
	"errors"
	"syscall"
)

func setMark(_ int) controlFunc {
	return func(string, string, syscall.RawConn) error {
		return errors.New("fwmark: linux only")
	}
}
