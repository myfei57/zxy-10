package batch

import (
	"edgelog/internal/store"
)

// Assemble builds a new batch record from the given lines without persisting
// it yet. The batch is only durable once Stage succeeds.
func (s *Service) Assemble(sourceID, topic string, lines []string, endOffset int64) (store.BatchRecord, error) {
	record := store.BatchRecord{
		ID:        NewID(),
		SourceID:  sourceID,
		Topic:     topic,
		Lines:     lines,
		EndOffset: endOffset,
		State:     store.BatchAssembled,
		CreatedAt: s.now(),
		UpdatedAt: s.now(),
	}
	return record, nil
}
