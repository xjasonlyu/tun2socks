package core

import (
	"sync"
)

const defaultBufferSize = 2 * 1024

var bufPool = sync.Pool{New: func() interface{} { return make([]byte, defaultBufferSize) }}

func newBytes(size int) []byte {
	if size <= defaultBufferSize {
		return bufPool.Get().([]byte)
	} else {
		return make([]byte, size)
	}
}

func freeBytes(b []byte) {
	if len(b) >= defaultBufferSize {
		bufPool.Put(b)
	}
}
