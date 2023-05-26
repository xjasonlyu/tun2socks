//go:build linux && !(amd64 || arm64)

package tun

import (
	"unsafe"
)

const (
	virtioNetHdrLen = int(unsafe.Sizeof(virtioNetHdr{}))
	offset          = virtioNetHdrLen + 0 /* NO_PI */
	defaultMTU      = 1500
)

// virtioNetHdr is defined in the kernel in include/uapi/linux/virtio_net.h. The
// kernel symbol is virtio_net_hdr.
type virtioNetHdr struct {
	flags      uint8
	gsoType    uint8
	hdrLen     uint16
	gsoSize    uint16
	csumStart  uint16
	csumOffset uint16
}
