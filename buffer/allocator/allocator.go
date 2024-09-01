package allocator

import (
	"errors"
	"math/bits"

	"github.com/xjasonlyu/tun2socks/v2/internal/pool"
)

// Allocator for incoming frames, optimized to prevent overwriting
// after zeroing.
type Allocator struct {
	buffers []*pool.Pool[[]byte]
}

// New initiates a []byte allocator for frames less than 65536 bytes,
// the waste(memory fragmentation) of space allocation is guaranteed
// to be no more than 50%.
func New() *Allocator {
	alloc := &Allocator{}
	alloc.buffers = make([]*pool.Pool[[]byte], 17) // 1B -> 64K
	for k := range alloc.buffers {
		i := k
		alloc.buffers[k] = pool.New(func() []byte {
			return make([]byte, 1<<uint32(i))
		})
	}
	return alloc
}

// Get gets a []byte from pool with most appropriate cap.
func (alloc *Allocator) Get(size int) []byte {
	if size <= 0 || size > 65536 {
		return nil
	}

	b := msb(size)
	if size == 1<<b {
		return alloc.buffers[b].Get()[:size]
	}

	return alloc.buffers[b+1].Get()[:size]
}

// Put returns a []byte to pool for future use,
// which the cap must be exactly 2^n.
func (alloc *Allocator) Put(buf []byte) error {
	b := msb(cap(buf))
	if cap(buf) == 0 || cap(buf) > 65536 || cap(buf) != 1<<b {
		return errors.New("allocator Put() incorrect buffer size")
	}

	alloc.buffers[b].Put(buf)
	return nil
}

// msb returns the pos of most significant bit.
func msb(size int) uint16 {
	return uint16(bits.Len32(uint32(size)) - 1)
}
