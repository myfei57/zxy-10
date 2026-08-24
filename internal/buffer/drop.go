package buffer

// Drop removes n lines from the front of the buffer and accounts for them in
// the drop counter. The counter must always reflect the lines actually
// discarded so flow control can react to backpressure.
func (b *RingBuffer) Drop(n int) {
	for i := 0; i < n && b.count > 0; i++ {
		b.lines[b.read] = ""
		b.read = (b.read + 1) % b.capacity
		b.count--
	}
}
