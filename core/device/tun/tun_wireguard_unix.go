//go:build unix

package tun

const (
	offset     = 4 /* 4 bytes TUN_PI */
	defaultMTU = 1500
)
