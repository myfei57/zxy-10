package verifycase

import (
	"testing"

	"edgelog/internal/batch"
	"edgelog/internal/dedup"
	"edgelog/internal/sink"
	"edgelog/internal/store"
	"edgelog/internal/tailer"
)

// TestSinkRetryNeverCachesErrorAsSuccess verifies a receiver 5xx always marks
// the batch failed instead of committing it.
func TestSinkRetryNeverCachesErrorAsSuccess(t *testing.T) {
	root := t.TempDir()
	st, err := store.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	checkpoints := store.NewCheckpointStore(st.FS)
	batchesStore := store.NewBatchStore(st.FS)
	tailerSvc := tailer.NewService(st.FS, checkpoints)
	batchSvc := batch.NewService(st.FS, batchesStore, checkpoints, dedup.NewDeduper(60), tailerSvc)
	registry := sink.NewRegistry(st.FS)
	receiver, err := registry.Register("receiver")
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.SetFailures(receiver.ID, 10); err != nil {
		t.Fatal(err)
	}
	record := store.BatchRecord{
		ID:       batch.NewID(),
		SourceID: "src",
		Topic:    "topic",
		Lines:    []string{"line"},
		State:    store.BatchAssembled,
	}
	staged, err := batchSvc.Stage(record)
	if err != nil {
		t.Fatal(err)
	}
	result, err := sink.Retry(batchSvc, registry, staged, receiver.ID, 3)
	if err != nil {
		t.Fatal(err)
	}
	if result.State == store.BatchCommitted {
		t.Fatalf("receiver error was cached as success: state = %s", result.State)
	}
}
