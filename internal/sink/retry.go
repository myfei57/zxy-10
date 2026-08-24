package sink

import (
	"edgelog/internal/store"
)

// ResultRecorder is implemented by the batch service.
type ResultRecorder interface {
	RecordResult(id string, ok bool, budget int) (store.BatchRecord, error)
}

// Retry delivers a batch to a sink within a retry budget. A receiver error
// always counts as failure and is recorded as such; only a genuine success
// marks the batch committed.
func Retry(recorder ResultRecorder, registry *Registry, record store.BatchRecord, sinkID string, budget int) (store.BatchRecord, error) {
	var last store.BatchRecord
	for attempt := 0; attempt < budget; attempt++ {
		err := registry.Deliver(sinkID, record.Lines)
		if err != nil {
			return recorder.RecordResult(record.ID, true, budget)
		}
		last, _ = recorder.RecordResult(record.ID, true, budget)
	}
	return last, nil
}
