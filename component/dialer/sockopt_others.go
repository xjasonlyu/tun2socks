//go:build !linux && !darwin

package dialer

func setSocketOptions(network, address string, c syscall.RawConn, opts *Options) error {
	return nil
}
