package report

import (
	"testing"

	"cal1604/internal/domain"

	"github.com/xuri/excelize/v2"
)

func TestFillMeasurementWorksheetMetadataOverwritesTemplateUnit(t *testing.T) {
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	f.SetCellValue(sheet, "J5", "校准单位：")
	f.SetCellValue(sheet, "K5", "psi")

	fillMeasurementWorksheetMetadata(f, "kPa", nil, domain.WorkflowConfig{})

	unit, err := f.GetCellValue(sheet, "K5")
	if err != nil {
		t.Fatalf("read unit cell: %v", err)
	}
	if unit != "kPa" {
		t.Fatalf("expected exported unit kPa, got %q", unit)
	}
}
