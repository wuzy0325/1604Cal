package report_test

import (
	"os"
	"path/filepath"
	"testing"

	"cal1604/internal/report"
)

func TestGetTemplatesParsesTemplateMetadata(t *testing.T) {
	templateDir := t.TempDir()
	for _, name := range []string{"6s.xlsx", "6m.xlsx", "README.txt", "bad-template.xlsx"} {
		if err := os.WriteFile(filepath.Join(templateDir, name), []byte("test"), 0o644); err != nil {
			t.Fatalf("write template file %s: %v", name, err)
		}
	}

	svc := report.NewService(templateDir)
	templates, err := svc.GetTemplates()
	if err != nil {
		t.Fatalf("GetTemplates: %v", err)
	}

	if len(templates) != 2 {
		t.Fatalf("expected 2 valid templates, got %d", len(templates))
	}

	seen := map[string]report.ReportTemplate{}
	for _, tpl := range templates {
		seen[tpl.Name] = tpl
	}

	if got, ok := seen["6s"]; !ok {
		t.Fatalf("expected template 6s to exist, got %v", templates)
	} else if got.Mode != "single" || got.PointCount != 6 {
		t.Fatalf("unexpected 6s template metadata: %+v", got)
	}

	if got, ok := seen["6m"]; !ok {
		t.Fatalf("expected template 6m to exist, got %v", templates)
	} else if got.Mode != "roundTrip" || got.PointCount != 6 {
		t.Fatalf("unexpected 6m template metadata: %+v", got)
	}
}

func TestMatchTemplateResolvesExpectedFilename(t *testing.T) {
	templateDir := t.TempDir()
	for _, name := range []string{"5s.xlsx", "5m.xlsx"} {
		if err := os.WriteFile(filepath.Join(templateDir, name), []byte("test"), 0o644); err != nil {
			t.Fatalf("write template file %s: %v", name, err)
		}
	}

	svc := report.NewService(templateDir)

	singlePath, err := svc.MatchTemplate(5, "single")
	if err != nil {
		t.Fatalf("MatchTemplate single: %v", err)
	}
	if filepath.Base(singlePath) != "5s.xlsx" {
		t.Fatalf("expected 5s.xlsx, got %s", singlePath)
	}

	roundTripPath, err := svc.MatchTemplate(5, "roundTrip")
	if err != nil {
		t.Fatalf("MatchTemplate roundTrip: %v", err)
	}
	if filepath.Base(roundTripPath) != "5m.xlsx" {
		t.Fatalf("expected 5m.xlsx, got %s", roundTripPath)
	}
}

func TestMatchTemplateReturnsErrorWhenMissing(t *testing.T) {
	templateDir := t.TempDir()
	svc := report.NewService(templateDir)

	if _, err := svc.MatchTemplate(4, "single"); err == nil {
		t.Fatal("expected missing template to return error")
	}
}
