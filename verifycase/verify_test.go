package verifycase

import (
	"testing"

	"edgelog/internal/config"
	"edgelog/internal/deadletter"
	"edgelog/internal/store"
)

// TestDeadletterRequeueUsesCurrentPartition verifies requeue resolves the
// live routing map instead of replaying the failure-time snapshot.
func TestDeadletterRequeueUsesCurrentPartition(t *testing.T) {
	root := t.TempDir()
	st, err := store.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	service := deadletter.NewService(st.FS)
	letter, err := service.Hold("batch-1", "src", "topic", "old-sink", 1, "budget exhausted")
	if err != nil {
		t.Fatal(err)
	}
	resolve := func(topic string) (config.RoutingRule, error) {
		return config.RoutingRule{Topic: topic, SinkID: "new-sink", Partition: 2}, nil
	}
	requeued, rule, err := service.Requeue(letter.ID, resolve)
	if err != nil {
		t.Fatal(err)
	}
	if !requeued.Requeued {
		t.Fatal("dead letter was not marked requeued")
	}
	if rule.Partition != 2 {
		t.Fatalf("requeue replayed a stale partition snapshot: %d", rule.Partition)
	}
}
