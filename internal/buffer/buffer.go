package buffer

import (
	"sync"
)

// Buffer is a simple type adapted from the built-in bytes.Buffer type.
type Buffer []byte

// global buffer pool - having an initial size gives a dramatic speedup
var bufPool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, 1024)
		return (*Buffer)(&b)
	},
}

// New creates a new buffer.
func New() *Buffer {
	return bufPool.Get().(*Buffer)
}

// Bytes returns the bytes in the buffer.
func (b *Buffer) Bytes() []byte {
	return *b
}

// Free returns buffers to the pool.
func (b *Buffer) Free() {
	// To reduce peak allocation, return only smaller buffers to the pool.
	const maxBufferSize = 16 << 10
	if cap(*b) <= maxBufferSize {
		*b = (*b)[:0]
		bufPool.Put(b)
	}
}

// Len returns the number of bytes currently in the buffer.
func (b *Buffer) Len() int {
	return len(*b)
}

// Reset clears the buffer.
func (b *Buffer) Reset() {
	*b = (*b)[:0]
}

// Write handles writing bytes to the buffer.
func (b *Buffer) Write(p []byte) (int, error) {
	*b = append(*b, p...)
	return len(p), nil
}

// Write string handles writing the string to the buffer.
func (b *Buffer) WriteString(s string) {
	*b = append(*b, s...)
}

// WriteByte handles writing a single byte to the buffer.
func (b *Buffer) WriteByte(c byte) error {
	*b = append(*b, c)
	return nil
}

// WritePosInt writes a non-negative integer to the buffer.
func (b *Buffer) WritePosInt(i int) {
	b.WritePosIntWidth(i, 0)
}

// WritePosIntWidth writes non-negative integer i to the buffer, padded on the left
// by zeroes to the given width.
//
// Use a width of 0 to omit padding.
func (b *Buffer) WritePosIntWidth(i, width int) {
	// Cheap integer to fixed-width decimal ASCII.
	// Copied from log/log.go.

	if i < 0 {
		panic("negative int")
	}

	// Assemble decimal in reverse order.
	var bb [20]byte
	bp := len(bb) - 1
	for i >= 10 || width > 1 {
		width--
		q := i / 10
		bb[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	bb[bp] = byte('0' + i)
	_, _ = b.Write(bb[bp:])
}

func (b *Buffer) String() string {
	return string(*b)
}
