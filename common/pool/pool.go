// Package pool provides a pool of []byte.
package pool

// Ref: github.com/Dreamacro/clash/common/pool

const (
	// MaxSegmentSize is the largest possible UDP datagram size.
	MaxSegmentSize = (1 << 16) - 1

	// io.Copy default buffer size is 32 KiB, but the maximum packet
	// size of vmess/shadowsocks is about 16 KiB, so define a buffer
	// of 20 KiB to reduce the memory of each TCP relay.
	RelayBufferSize = 20 << 10
)

// Get gets a []byte from default allocator with most appropriate cap.
func Get(size int) []byte {
	return _allocator.Get(size)
}

// Put returns a []byte to default allocator for future use.
func Put(buf []byte) error {
	return _allocator.Put(buf)
}
