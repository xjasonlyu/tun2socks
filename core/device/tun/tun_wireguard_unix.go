//go:build unix

package tun

//nolint:all
const (
	offset     = 4 /* 4 bytes TUN_PI */
	defaultMTU = 1500
)
