package pool

import (
	"bytes"
	"sync"
)

var _bufferPool = sync.Pool{New: func() any { return &bytes.Buffer{} }}

func GetBuffer() *bytes.Buffer {
	return _bufferPool.Get().(*bytes.Buffer)
}

func PutBuffer(buf *bytes.Buffer) {
	buf.Reset()
	_bufferPool.Put(buf)
}
