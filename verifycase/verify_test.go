package verifycase

import (
	"os"
	"testing"

	"edgelog/internal/config"
	"edgelog/internal/store"
	"edgelog/internal/tailer"
)

// TestPipelineSwitchesAfterConfigDurable verifies the new pipeline attaches
// only after its config is durably stored.
func TestPipelineSwitchesAfterConfigDurable(t *testing.T) {
	root := t.TempDir()
	st, err := store.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	pipelines := tailer.NewPipelineManager()
	switchSvc := config.NewSwitchService(st.FS, pipelines)
	first := config.PipelineConfig{SourceID: "src", BatchSize: 10, SinkIDs: []string{"s1"}}
	if _, err := switchSvc.Switch("src", first); err != nil {
		t.Fatal(err)
	}
	block := switchSvc.Path("src")
	if err := os.Chmod(block, 0o444); err != nil {
		t.Fatal(err)
	}
	second := config.PipelineConfig{SourceID: "src", BatchSize: 20, SinkIDs: []string{"s2"}}
	if _, err := switchSvc.Switch("src", second); err == nil {
		t.Fatal("switch should fail when the config cannot be written")
	}
	active, ok := pipelines.Active("src")
	if !ok {
		t.Fatal("pipeline is missing after a failed switch")
	}
	if active.BatchSize != 10 {
		t.Fatalf("pipeline switched before config durable: batch_size = %d", active.BatchSize)
	}
}
