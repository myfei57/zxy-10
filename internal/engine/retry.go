package engine

import (
	"edgelog/internal/sink"
	"edgelog/internal/store"
)

func (e *Engine) sinkRetry(record store.BatchRecord, sinkID string) (store.BatchRecord, error) {
	return sink.Retry(e.Batches, e.Sinks, record, sinkID, e.cfg.RetryBudget)
}
