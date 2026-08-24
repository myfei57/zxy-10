package verifycase

import (
	"os"
	"testing"

	"edgelog/internal/batch"
	"edgelog/internal/dedup"
	"edgelog/internal/store"
	"edgelog/internal/tailer"
)

// TestBatchCommitsAfterStageDurable verifies the consumption cursor commits
// only after the staged batch file is durably written.
func TestBatchCommitsAfterStageDurable(t *testing.T) {
	root := t.TempDir()
	st, err := store.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	checkpoints := store.NewCheckpointStore(st.FS)
	batchesStore := store.NewBatchStore(st.FS)
	tailerSvc := tailer.NewService(st.FS, checkpoints)
	batchSvc := batch.NewService(st.FS, batchesStore, checkpoints, dedup.NewDeduper(60), tailerSvc)

	record, err := batchSvc.Assemble("src", "topic", []string{"a", "b"}, 42)
	if err != nil {
		t.Fatal(err)
	}
	block := batchesStore.Path(record.ID)
	if err := os.WriteFile(block, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(block, 0o444); err != nil {
		t.Fatal(err)
	}
	if _, err := batchSvc.Commit("src", record, 42, 2); err == nil {
		t.Fatal("commit should fail when the staged batch cannot be written")
	}
	got, err := tailerSvc.Current("src")
	if err != nil {
		t.Fatal(err)
	}
	if got.Offset != 0 {
		t.Fatalf("cursor committed before batch durable: offset = %d", got.Offset)
	}
}
