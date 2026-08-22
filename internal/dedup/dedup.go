package dedup

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// Deduper rejects repeated log lines inside a sliding window. A line whose
// fingerprint is already in the open window is a duplicate and must not be
// ingested twice.
type Deduper struct {
	windows       map[string]map[string]time.Time
	windowSeconds int
	clock         func() time.Time
}

// NewDeduper creates a deduper with the given window.
func NewDeduper(windowSeconds int) *Deduper {
	return &Deduper{
		windows:       make(map[string]map[string]time.Time),
		windowSeconds: windowSeconds,
		clock:         time.Now,
	}
}

func fingerprint(sourceID, line string) string {
	sum := sha256.Sum256([]byte(sourceID + "\x00" + line))
	return hex.EncodeToString(sum[:])
}

// Accept reports whether a line may be ingested: false means it is a
// duplicate inside the open window.
func (d *Deduper) Accept(sourceID, line string) bool {
	window := d.windows[sourceID]
	if window == nil {
		window = make(map[string]time.Time)
		d.windows[sourceID] = window
	}
	now := d.clock()
	fp := fingerprint(sourceID, line)
	if seenAt, ok := window[fp]; ok && now.Sub(seenAt) <= time.Duration(d.windowSeconds)*time.Second {
		return false
	}
	window[fp] = now
	return true
}
