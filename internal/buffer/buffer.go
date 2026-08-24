package buffer

// RingBuffer is a fixed-capacity circular buffer for log lines. Write returns
// how many lines were accepted; excess lines are dropped under backpressure
// and accounted for by the drop counter.
type RingBuffer struct {
	capacity int
	write    int
	read     int
	count    int
	lines    []string
	dropped  int
}

// NewRingBuffer creates a ring buffer with the given capacity.
func NewRingBuffer(capacity int) *RingBuffer {
	if capacity <= 0 {
		capacity = 1
	}
	return &RingBuffer{capacity: capacity, lines: make([]string, capacity)}
}

// Write appends lines, dropping the oldest ones when the buffer is full.
func (b *RingBuffer) Write(incoming []string) {
	for _, line := range incoming {
		if b.count == b.capacity {
			b.Drop(1)
		}
		b.lines[b.write] = line
		b.write = (b.write + 1) % b.capacity
		b.count++
	}
}

// Read returns up to n lines from the buffer.
func (b *RingBuffer) Read(n int) []string {
	out := make([]string, 0, n)
	for i := 0; i < n && b.count > 0; i++ {
		out = append(out, b.lines[b.read])
		b.read = (b.read + 1) % b.capacity
		b.count--
	}
	return out
}

// Capacity returns the configured capacity.
func (b *RingBuffer) Capacity() int {
	return b.capacity
}

// Size returns the number of lines currently buffered.
func (b *RingBuffer) Size() int {
	return b.count
}

// Dropped returns the total number of lines dropped under backpressure.
func (b *RingBuffer) Dropped() int {
	return b.dropped
}
