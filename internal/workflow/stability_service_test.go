package workflow_test

import (
	"testing"
	"time"

	"cal1604/internal/workflow"
)

func TestStabilityAccumulatorResetsOnDrift(t *testing.T) {
	acc := workflow.NewStabilityAccumulator(0.2, 2*time.Second)

	stable, elapsed := acc.AddSample(10, 10.1, time.Second)
	if stable {
		t.Fatal("expected not stable after first sample")
	}
	if elapsed != time.Second {
		t.Fatalf("expected 1s elapsed, got %s", elapsed)
	}

	stable, elapsed = acc.AddSample(10, 10.6, time.Second)
	if stable {
		t.Fatal("expected drift sample to be unstable")
	}
	if elapsed != 0 {
		t.Fatalf("expected elapsed reset to 0, got %s", elapsed)
	}
}
