package dedup

// Close closes the dedup window of a source. It is invoked by the batch
// commit boundary; until then the window stays open so a delayed replay of
// an uncommitted line is still recognised as a duplicate.
func (d *Deduper) Close(sourceID string) {
	delete(d.windows, sourceID)
}
