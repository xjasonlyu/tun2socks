//go:build !linux && !darwin && !windows

package dialer

import "syscall"

func setSocketOptions(network, address string, c syscall.RawConn, opts *Options) error {
	return nil
}
