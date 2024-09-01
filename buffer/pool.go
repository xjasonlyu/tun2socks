// Package buffer provides a pool of []byte.
package buffer

import (
	"github.com/xjasonlyu/tun2socks/v2/buffer/allocator"
)

const (
	// MaxSegmentSize is the largest possible UDP datagram size.
	MaxSegmentSize = (1 << 16) - 1

	// RelayBufferSize is the default buffer size for TCP relays.
	// io.Copy default buffer size is 32 KiB, but the maximum packet
	// size of vmess/shadowsocks is about 16 KiB, so define a buffer
	// of 20 KiB to reduce the memory of each TCP relay.
	RelayBufferSize = 20 << 10
)

var _allocator = allocator.New()

// Get gets a []byte from default allocator with most appropriate cap.
func Get(size int) []byte {
	return _allocator.Get(size)
}

// Put returns a []byte to default allocator for future use.
func Put(buf []byte) error {
	return _allocator.Put(buf)
}
