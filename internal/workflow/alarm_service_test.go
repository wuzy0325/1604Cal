package workflow_test

import (
	"testing"

	"cal1604/internal/workflow"
)

func TestAlarmDecisionAllowsContinueOrRetryOnly(t *testing.T) {
	svc := workflow.NewAlarmService()

	if err := svc.ValidateDecision("continue"); err != nil {
		t.Fatalf("continue should be valid, got %v", err)
	}

	if err := svc.ValidateDecision("retry"); err != nil {
		t.Fatalf("retry should be valid, got %v", err)
	}

	if err := svc.ValidateDecision("skip"); err == nil {
		t.Fatal("expected invalid decision to fail")
	}
}

func TestAlarmEvaluateDeviation(t *testing.T) {
	svc := workflow.NewAlarmService()

	result := svc.Evaluate(50, 52.35, 0.5)
	if !result.Triggered {
		t.Fatal("expected alarm to be triggered")
	}

	if result.DeviationPercent <= 0.5 {
		t.Fatalf("expected deviation > 0.5%%, got %.3f", result.DeviationPercent)
	}
}
