package report_test

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"cal1604/internal/domain"
	"cal1604/internal/report"

	"github.com/xuri/excelize/v2"
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

// buildRoundTripPoints 构造一组带正程+回程的压力点，用于回归测试。
// 每个点写入 numChannels 个采样值 = baseValue + ch*0.1。
// skipBackwardTargets 中列出的标准压力对应的回程点会被标记为 pending 以模拟"未完成"。
func buildRoundTripPoints(t *testing.T, standards []float64, numChannels int, skipBackwardTargets map[float64]struct{}) []domain.PressurePoint {
	t.Helper()
	points := make([]domain.PressurePoint, 0, len(standards)*2)
	for i, std := range standards {
		samples := make([]float64, numChannels)
		for ch := 0; ch < numChannels; ch++ {
			samples[ch] = std + 0.001 + float64(ch)*0.0001
		}
		points = append(points, domain.PressurePoint{
			Index:          i + 1,
			TargetPressure: std,
			Direction:      "forward",
			Status:         domain.PointStatusCompleted,
			CollectedData:  samples,
		})
	}
	// 回程按 max → min 顺序
	for i := len(standards) - 1; i >= 0; i-- {
		std := standards[i]
		pt := domain.PressurePoint{
			Index:          len(standards) + (len(standards) - i),
			TargetPressure: std,
			Direction:      "backward",
			Status:         domain.PointStatusCompleted,
		}
		if _, skip := skipBackwardTargets[std]; skip {
			pt.Status = domain.PointStatusPending
		} else {
			samples := make([]float64, numChannels)
			for ch := 0; ch < numChannels; ch++ {
				samples[ch] = std + 0.5 + float64(ch)*0.0001
			}
			pt.CollectedData = samples
		}
		points = append(points, pt)
	}
	return points
}

// TestExportMeasurementReportRoundTripFillsReturnStroke 回归测试：
// 修复 Return stroke 全 0 的缺陷，确保 fallback 工作簿在回程模式下
// 正确按标准压力对齐写入 B/C 列。
func TestExportMeasurementReportRoundTripFillsReturnStroke(t *testing.T) {
	svc := report.NewService("") // 不指定模板目录 → 走 fallback
	numChannels := 16
	standards := []float64{0, 10, 20, 30, 40}
	points := buildRoundTripPoints(t, standards, numChannels, nil)

	outPath := filepath.Join(t.TempDir(), "round-trip.xlsx")
	cfg := domain.WorkflowConfig{
		PointCount:     len(standards),
		PressureMode:   domain.PressureModeRoundTrip,
		AverageCount:   1,
		MinPressure:    0,
		MaxPressure:    40,
		PrecisionLevel: 0.001,
	}
	if _, err := svc.ExportMeasurementReport(context.Background(), points, cfg, outPath, "kPa"); err != nil {
		t.Fatalf("ExportMeasurementReport: %v", err)
	}

	f, err := excelize.OpenFile(outPath)
	if err != nil {
		t.Fatalf("open exported file: %v", err)
	}
	defer f.Close()

	sheet := "校准结果"
	// fallback 布局：第一通道块 blockStart=7, dataStartRow=10。
	const dataStartRow = 10
	for row := 0; row < len(standards); row++ {
		excelRow := dataStartRow + row
		std := standards[row]

		gotStd, _ := f.GetCellValue(sheet, "A"+itoa(excelRow))
		if gotStd == "" {
			t.Fatalf("A%d empty, expect standard pressure %.2f", excelRow, std)
		}

		// B 列：正程显示值，应≈ std + 0.001（第 1 通道）
		gotForward, _ := f.GetCellValue(sheet, "B"+itoa(excelRow))
		if gotForward == "" {
			t.Errorf("B%d empty, expect forward value", excelRow)
		}

		// C 列：回程显示值，应≈ std + 0.5（第 1 通道）；本测试核心断言。
		gotReturn, _ := f.GetCellValue(sheet, "C"+itoa(excelRow))
		if gotReturn == "" || gotReturn == "0" {
			t.Errorf("C%d expected non-zero return stroke for standard=%.2f, got %q", excelRow, std, gotReturn)
		}
	}
}

// TestExportMeasurementReportSkipsMissingBackwardPoints 验证：
// 当部分回程点未完成时，C 列应仅在已完成行写入，不发生静默错位。
func TestExportMeasurementReportSkipsMissingBackwardPoints(t *testing.T) {
	svc := report.NewService("")
	numChannels := 16
	standards := []float64{0, 10, 20, 30, 40}
	// 故意跳过 standard=20 的回程点
	skip := map[float64]struct{}{20: {}}
	points := buildRoundTripPoints(t, standards, numChannels, skip)

	outPath := filepath.Join(t.TempDir(), "round-trip-missing.xlsx")
	cfg := domain.WorkflowConfig{
		PointCount:     len(standards),
		PressureMode:   domain.PressureModeRoundTrip,
		AverageCount:   1,
		MinPressure:    0,
		MaxPressure:    40,
		PrecisionLevel: 0.001,
	}
	if _, err := svc.ExportMeasurementReport(context.Background(), points, cfg, outPath, "kPa"); err != nil {
		t.Fatalf("ExportMeasurementReport: %v", err)
	}

	f, err := excelize.OpenFile(outPath)
	if err != nil {
		t.Fatalf("open exported file: %v", err)
	}
	defer f.Close()

	sheet := "校准结果"
	const dataStartRow = 10
	// 标准压力 20 对应 row offset=2 → excelRow=12，C 列应保持空白；
	// 其他行的 C 列应有值。
	for row, std := range standards {
		excelRow := dataStartRow + row
		got, _ := f.GetCellValue(sheet, "C"+itoa(excelRow))
		if std == 20 {
			if got != "" {
				t.Errorf("C%d (standard=%.2f) expected empty for skipped backward point, got %q", excelRow, std, got)
			}
		} else {
			if got == "" || got == "0" {
				t.Errorf("C%d (standard=%.2f) expected non-zero return value, got %q", excelRow, std, got)
			}
		}
	}
}

// itoa 把 int 转字符串，仅是 strconv.Itoa 的本地别名以保持调用点简洁。
func itoa(n int) string {
	return strconv.Itoa(n)
}
