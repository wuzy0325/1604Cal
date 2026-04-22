package report_test

import (
	"testing"

	"cal1604/internal/report"
)

func TestSelectTemplate(t *testing.T) {
	got, err := report.SelectTemplate(5, "single")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "5s.xlsx" {
		t.Fatalf("expected 5s.xlsx, got %s", got)
	}
}

func TestSelectTemplateRejectsInvalidPointCount(t *testing.T) {
	if _, err := report.SelectTemplate(1, "single"); err == nil {
		t.Fatal("expected point count validation error")
	}
}

func TestSelectTemplateRejectsPointCountAboveSix(t *testing.T) {
	if _, err := report.SelectTemplate(7, "single"); err == nil {
		t.Fatal("expected point count above six to be rejected")
	}
}
