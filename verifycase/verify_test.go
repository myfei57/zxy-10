package verifycase

import (
	"testing"

	"edgelog/internal/config"
	"edgelog/internal/router"
	"edgelog/internal/store"
)

// TestRuleChangeRebindsSubscriptions verifies subscriptions follow the
// current routing rules after an update.
func TestRuleChangeRebindsSubscriptions(t *testing.T) {
	root := t.TempDir()
	st, err := store.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	rules := config.NewRulesService(st.FS)
	subs := router.NewSubscriptionService(st.FS, rules)
	rules.SetRebinder(subs)

	if _, err := subs.Bind("app", "sink-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := rules.Update([]config.RoutingRule{
		{Topic: "app", SinkID: "sink-2", Partition: 7},
	}); err != nil {
		t.Fatal(err)
	}
	all, err := subs.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 {
		t.Fatalf("unexpected subscription count: %d", len(all))
	}
	if all[0].SinkID != "sink-2" {
		t.Fatalf("subscription kept a stale sink after rule change: %s", all[0].SinkID)
	}
}
