package bufferpool

import (
	"bytes"

	"github.com/xjasonlyu/tun2socks/v2/internal/pool"
)

const _size = 1024 // by default, create 1 KiB buffers

var _pool = pool.New(func() *bytes.Buffer {
	buf := &bytes.Buffer{}
	buf.Grow(_size)
	return buf
})

func Get() *bytes.Buffer {
	buf := _pool.Get()
	buf.Reset()
	return buf
}

func Put(b *bytes.Buffer) {
	_pool.Put(b)
}
