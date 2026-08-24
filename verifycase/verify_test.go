package verifycase

import (
	"os"
	"testing"

	"edgelog/internal/store"
	"edgelog/internal/tailer"
)

// TestTailerCursorAdvancesAfterCheckpointDurable verifies the read cursor
// only advances after the checkpoint is durably stored.
func TestTailerCursorAdvancesAfterCheckpointDurable(t *testing.T) {
	root := t.TempDir()
	st, err := store.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	checkpoints := store.NewCheckpointStore(st.FS)
	tailerSvc := tailer.NewService(st.FS, checkpoints)

	current, err := tailerSvc.Current("src")
	if err != nil {
		t.Fatal(err)
	}
	advanced, err := tailerSvc.Advance(current, tailer.Cursor{SourceID: "src", Offset: 10, Line: 1})
	if err != nil {
		t.Fatal(err)
	}
	block := checkpoints.Path("src")
	if err := os.Chmod(block, 0o444); err != nil {
		t.Fatal(err)
	}
	if _, err := tailerSvc.Advance(advanced, tailer.Cursor{SourceID: "src", Offset: 20, Line: 2}); err == nil {
		t.Fatal("advance should fail when the checkpoint cannot be written")
	}
	got, err := tailerSvc.Current("src")
	if err != nil {
		t.Fatal(err)
	}
	if got.Offset != 10 {
		t.Fatalf("cursor advanced before checkpoint durable: offset = %d", got.Offset)
	}
}
