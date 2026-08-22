package verifycase

import (
	"testing"

	"edgelog/internal/quota"
	"edgelog/internal/router"
)

// TestQuotaRecoveryResetsLimiter verifies a recovered quota window clears the
// flow-control limiter.
func TestQuotaRecoveryResetsLimiter(t *testing.T) {
	flow := router.NewFlowController()
	quotaSvc := quota.NewServiceWithLimiter(100, 60, flow)
	flow.SetThrottled("src", true)
	quotaSvc.Recover("src")
	if flow.IsThrottled("src") {
		t.Fatal("limiter stayed throttled after quota recovery")
	}
}
