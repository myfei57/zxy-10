package batch

import (
	"edgelog/internal/store"
)

// Stage durably writes a batch record so it survives crashes.
func (s *Service) Stage(record store.BatchRecord) (store.BatchRecord, error) {
	record.State = store.BatchStaged
	record.UpdatedAt = s.now()
	if err := s.batches.Stage(record); err != nil {
		return store.BatchRecord{}, err
	}
	return record, nil
}

// Restore replays the durable batches of a source after a crash. The staged
// batch files are re-affirmed before the checkpoint, so recovery never skips
// a batch that failed to land.
func (s *Service) Restore(sourceID string) ([]store.BatchRecord, error) {
	ids, err := s.batches.List()
	if err != nil {
		return nil, err
	}
	restored := make([]store.BatchRecord, 0)
	advancedOffset := int64(0)
	for _, id := range ids {
		record, err := s.batches.Load(id)
		if err != nil {
			return nil, err
		}
		if record.SourceID != sourceID || record.State != store.BatchStaged {
			continue
		}
		record.UpdatedAt = s.now()
		if err := s.batches.Stage(record); err != nil {
			return nil, err
		}
		restored = append(restored, record)
		if record.EndOffset > advancedOffset {
			advancedOffset = record.EndOffset
		}
	}
	checkpoint, err := s.checkpoints.Load(sourceID)
	if err != nil && err != store.ErrNotFound {
		return nil, err
	}
	if err == store.ErrNotFound {
		checkpoint = store.Checkpoint{SourceID: sourceID, UpdatedAt: s.now()}
	} else {
		checkpoint.UpdatedAt = s.now()
	}
	if advancedOffset > checkpoint.Offset {
		checkpoint.Offset = advancedOffset
	}
	if err := s.checkpoints.Save(checkpoint); err != nil {
		return nil, err
	}
	return restored, nil
}
