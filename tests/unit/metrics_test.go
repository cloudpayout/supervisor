package unit

import (
	"os"
	"testing"

	"github.com/cloudpayout/supervisor/internal/metrics"
)

func TestAggregateSelfAndMissingPID(t *testing.T) {
	self := metrics.Aggregate(os.Getpid(), 0)
	if !self.OK {
		t.Fatal("expected metrics for current process")
	}
	if self.MemoryRSS == 0 {
		t.Fatal("expected non-zero RSS")
	}
	missing := metrics.Aggregate(1<<30, 0)
	if missing.OK {
		t.Fatal("expected unavailable metrics for a missing pid")
	}
}
