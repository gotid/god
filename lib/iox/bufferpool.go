package iox

import (
	"bytes"
	"sync"
)

// BufferPool 是一个 bytes.Buffer 对象的缓冲池。
type BufferPool struct {
	capability int
	pool       *sync.Pool
}

// NewBufferPool 返回一个 BufferPool。
func NewBufferPool(capacity int) *BufferPool {
	return &BufferPool{
		capability: capacity,
		pool: &sync.Pool{
			New: func() any {
				return new(bytes.Buffer)
			},
		},
	}
}

// Get 从缓冲池获取 bytes.Buffer。
func (bp *BufferPool) Get() *bytes.Buffer {
	buf := bp.pool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

func (bp *BufferPool) Put(buf *bytes.Buffer) {
	if buf.Cap() < bp.capability {
		bp.pool.Put(buf)
	}
}
