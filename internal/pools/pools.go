package pools

import (
	"sync"
)

// BufferPool provides a pool of reusable byte buffers
type BufferPool struct {
	pool sync.Pool
}

// NewBufferPool creates a new buffer pool with buffers of the specified size
func NewBufferPool(size int) *BufferPool {
	return &BufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				buf := make([]byte, size)
				return &buf
			},
		},
	}
}

// Get retrieves a buffer from the pool
func (p *BufferPool) Get() *[]byte {
	return p.pool.Get().(*[]byte)
}

// Put returns a buffer to the pool for reuse
func (p *BufferPool) Put(buf *[]byte) {
	if buf != nil && cap(*buf) > 0 {
		// Clear the buffer before returning to pool
		*buf = (*buf)[:cap(*buf)]
		for i := range *buf {
			(*buf)[i] = 0
		}
		p.pool.Put(buf)
	}
}

// Common buffer sizes
var (
	// SmallBufferPool for small allocations (4KB)
	SmallBufferPool = NewBufferPool(4 * 1024)

	// MediumBufferPool for medium allocations (64KB)
	MediumBufferPool = NewBufferPool(64 * 1024)

	// LargeBufferPool for large allocations (1MB)
	LargeBufferPool = NewBufferPool(1024 * 1024)
)

// ReportPool provides object pooling for Report structures
type ReportPool struct {
	pool sync.Pool
}

// NewReportPool creates a pool for Report objects
func NewReportPool(factory func() interface{}) *ReportPool {
	return &ReportPool{
		pool: sync.Pool{
			New: factory,
		},
	}
}

// Get retrieves a Report from the pool
func (p *ReportPool) Get() interface{} {
	return p.pool.Get()
}

// Put returns a Report to the pool
func (p *ReportPool) Put(report interface{}) {
	p.pool.Put(report)
}