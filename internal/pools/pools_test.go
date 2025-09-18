package pools

import (
	"testing"
)

func TestNewBufferPool(t *testing.T) {
	size := 1024
	pool := NewBufferPool(size)

	if pool == nil {
		t.Fatal("NewBufferPool() returned nil")
	}

	// Get a buffer from the pool
	buf := pool.Get()
	if buf == nil {
		t.Fatal("Get() returned nil")
	}

	if len(*buf) != size {
		t.Errorf("Expected buffer size %d, got %d", size, len(*buf))
	}
}

func TestBufferPoolGetAndPut(t *testing.T) {
	pool := NewBufferPool(256)

	// Get a buffer
	buf1 := pool.Get()
	if buf1 == nil {
		t.Fatal("First Get() returned nil")
	}

	// Modify the buffer
	(*buf1)[0] = 42
	(*buf1)[1] = 43

	// Return it to the pool
	pool.Put(buf1)

	// Get another buffer (should be the same one)
	buf2 := pool.Get()
	if buf2 == nil {
		t.Fatal("Second Get() returned nil")
	}

	// Buffer should be cleared
	if (*buf2)[0] != 0 || (*buf2)[1] != 0 {
		t.Error("Buffer not cleared when returned to pool")
	}
}

func TestBufferPoolPutNil(t *testing.T) {
	pool := NewBufferPool(128)

	// Putting nil should not panic
	pool.Put(nil)

	// Pool should still work
	buf := pool.Get()
	if buf == nil {
		t.Fatal("Get() returned nil after Put(nil)")
	}
}

func TestBufferPoolConcurrency(t *testing.T) {
	pool := NewBufferPool(512)

	// Run concurrent operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				buf := pool.Get()
				if buf == nil {
					t.Error("Get() returned nil in concurrent test")
				}
				// Use the buffer
				(*buf)[0] = byte(j)
				// Return it
				pool.Put(buf)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestSmallBufferPool(t *testing.T) {
	buf := SmallBufferPool.Get()
	if buf == nil {
		t.Fatal("SmallBufferPool.Get() returned nil")
	}

	expectedSize := 4 * 1024 // 4KB
	if len(*buf) != expectedSize {
		t.Errorf("Expected small buffer size %d, got %d", expectedSize, len(*buf))
	}

	SmallBufferPool.Put(buf)
}

func TestMediumBufferPool(t *testing.T) {
	buf := MediumBufferPool.Get()
	if buf == nil {
		t.Fatal("MediumBufferPool.Get() returned nil")
	}

	expectedSize := 64 * 1024 // 64KB
	if len(*buf) != expectedSize {
		t.Errorf("Expected medium buffer size %d, got %d", expectedSize, len(*buf))
	}

	MediumBufferPool.Put(buf)
}

func TestLargeBufferPool(t *testing.T) {
	buf := LargeBufferPool.Get()
	if buf == nil {
		t.Fatal("LargeBufferPool.Get() returned nil")
	}

	expectedSize := 1024 * 1024 // 1MB
	if len(*buf) != expectedSize {
		t.Errorf("Expected large buffer size %d, got %d", expectedSize, len(*buf))
	}

	LargeBufferPool.Put(buf)
}

func TestNewReportPool(t *testing.T) {
	factoryCalled := false
	factory := func() interface{} {
		factoryCalled = true
		return struct{ ID int }{ID: 42}
	}

	pool := NewReportPool(factory)
	if pool == nil {
		t.Fatal("NewReportPool() returned nil")
	}

	// Get an object from the pool
	obj := pool.Get()
	if obj == nil {
		t.Fatal("Get() returned nil")
	}

	if !factoryCalled {
		t.Error("Factory function was not called")
	}

	// Check the object
	report, ok := obj.(struct{ ID int })
	if !ok {
		t.Error("Get() returned wrong type")
	}

	if report.ID != 42 {
		t.Errorf("Expected ID 42, got %d", report.ID)
	}
}

func TestReportPoolGetAndPut(t *testing.T) {
	type TestReport struct {
		Value string
	}

	pool := NewReportPool(func() interface{} {
		return &TestReport{}
	})

	// Get an object
	obj1 := pool.Get()
	report1, ok := obj1.(*TestReport)
	if !ok {
		t.Fatal("Get() returned wrong type")
	}

	// Modify it
	report1.Value = "test"

	// Return it to the pool
	pool.Put(obj1)

	// Get another object (might be the same one)
	obj2 := pool.Get()
	if obj2 == nil {
		t.Fatal("Second Get() returned nil")
	}

	// Note: We can't guarantee it's the same object due to sync.Pool behavior,
	// but it should be the correct type
	_, ok = obj2.(*TestReport)
	if !ok {
		t.Error("Second Get() returned wrong type")
	}
}

func TestReportPoolConcurrency(t *testing.T) {
	type TestReport struct {
		ID int
	}

	nextID := 0
	pool := NewReportPool(func() interface{} {
		nextID++
		return &TestReport{ID: nextID}
	})

	// Run concurrent operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				obj := pool.Get()
				if obj == nil {
					t.Error("Get() returned nil in concurrent test")
				}
				// Use the object
				report, ok := obj.(*TestReport)
				if !ok {
					t.Error("Get() returned wrong type in concurrent test")
				}
				// Simulate some work
				report.ID = j
				// Return it
				pool.Put(obj)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func BenchmarkBufferPoolGetPut(b *testing.B) {
	pool := NewBufferPool(4096)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf := pool.Get()
		(*buf)[0] = byte(i)
		pool.Put(buf)
	}
}

func BenchmarkBufferPoolParallel(b *testing.B) {
	pool := NewBufferPool(4096)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := pool.Get()
			(*buf)[0] = 1
			pool.Put(buf)
		}
	})
}

func BenchmarkDirectAllocation(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf := make([]byte, 4096)
		buf[0] = byte(i)
	}
}