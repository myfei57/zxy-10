package batch

import (
	"edgelog/internal/store"
	"edgelog/internal/tailer"
)

// Commit persists a batch durably, then advances the source cursor and closes
// the dedup window. The order matters: the batch must land before the cursor
// moves and before the dedup window closes.
func (s *Service) Commit(sourceID string, record store.BatchRecord, nextOffset int64, nextLine int) (store.BatchRecord, error) {
	current, err := s.advancer.Current(sourceID)
	if err != nil {
		return store.BatchRecord{}, err
	}
	// Persist the batch before moving the cursor. If this fails the cursor
	// must stay put so the next tick re-reads and re-assembles these lines;
	// advancing first would drop the batch between a forwarded cursor and a
	// missing batch file with no record left to forward or dead-letter.
	staged, err := s.Stage(record)
	if err != nil {
		return store.BatchRecord{}, err
	}
	next := tailer.Cursor{SourceID: sourceID, Offset: nextOffset, Line: nextLine}
	if _, err := s.advancer.Advance(current, next); err != nil {
		return store.BatchRecord{}, err
	}
	if s.dedup != nil {
		s.dedup.Close(sourceID)
	}
	staged.State = store.BatchCommitted
	staged.UpdatedAt = s.now()
	if err := s.batches.Stage(staged); err != nil {
		return store.BatchRecord{}, err
	}
	if s.audit != nil {
		_, _ = s.audit.Record("engine", "batch.commit", "batch", staged.ID, staged.SourceID)
	}
	return staged, nil
}
