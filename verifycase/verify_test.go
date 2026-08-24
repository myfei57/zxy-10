package verifycase

import (
	"testing"

	"edgelog/internal/buffer"
)

// TestBufferDropCounterStaysInSync verifies the drop counter reflects every
// line actually discarded under backpressure.
func TestBufferDropCounterStaysInSync(t *testing.T) {
	buf := buffer.NewRingBuffer(2)
	buf.Write([]string{"a", "b", "c", "d", "e"})
	stats := buf.Snapshot()
	if stats.Dropped != 3 {
		t.Fatalf("drop counter out of sync: dropped = %d, want 3", stats.Dropped)
	}
	if stats.Size != 2 {
		t.Fatalf("buffer size = %d, want 2", stats.Size)
	}
}
