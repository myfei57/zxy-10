package buffer

// Stats is a point-in-time snapshot of buffer accounting.
type Stats struct {
	Capacity int `json:"capacity"`
	Size     int `json:"size"`
	Dropped  int `json:"dropped"`
}

// Snapshot returns the current buffer accounting.
func (b *RingBuffer) Snapshot() Stats {
	return Stats{Capacity: b.capacity, Size: b.count, Dropped: b.dropped}
}
