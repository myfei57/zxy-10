package batch

import (
	"edgelog/internal/store"
)

// RecordResult updates the delivery state of a batch. Only genuine success
// may move a batch to committed; every failure keeps it failed or moves it to
// dead after the retry budget.
func (s *Service) RecordResult(id string, ok bool, budget int) (store.BatchRecord, error) {
	record, err := s.batches.Load(id)
	if err != nil {
		return store.BatchRecord{}, err
	}
	if ok {
		record.State = store.BatchCommitted
	} else {
		record.Retries++
		if record.Retries >= budget {
			record.State = store.BatchDead
		} else {
			record.State = store.BatchFailed
		}
	}
	record.UpdatedAt = s.now()
	if err := s.batches.Stage(record); err != nil {
		return store.BatchRecord{}, err
	}
	return record, nil
}
