package report

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"cal1604/internal/application/calibration"
)

// Service 封装报告模板路径拼装逻辑与报告导出。
type Service struct {
	templateDir string
}

// ReportTemplate 描述一个可用报告模板。
type ReportTemplate struct {
	Name       string `json:"name"`
	PointCount int    `json:"pointCount"`
	Mode       string `json:"mode"`
	Path       string `json:"path"`
}

// NewService 创建报告服务。
func NewService(templateDir string) *Service {
	return &Service{templateDir: templateDir}
}

// ResolveTemplatePath 解析模板绝对路径。
func (s *Service) ResolveTemplatePath(points int, mode string) (string, error) {
	return s.MatchTemplate(points, mode)
}

// ExportReport 根据 CalibrationSession 生成校准报告并保存到 outputPath。
// 优先使用模板文件填充数据，无模板时创建默认工作簿。
func (s *Service) ExportReport(ctx context.Context, session *calibration.CalibrationSession, outputPath string) error {
	if session == nil {
		return fmt.Errorf("calibration session is nil")
	}

	// 收集标准压力值（仅正程）
	standardValues := make([]float64, 0, len(session.Points))
	for _, p := range session.Points {
		if p.Direction == "" || p.Direction == "forward" {
			standardValues = append(standardValues, p.TargetPressure)
		}
	}

	// 收集通道数据
	channels := collectChannelData(session)

	// 确定压力单位
	unit := "kPa"

	// 尝试加载模板
	templatePath, _ := s.ResolveTemplatePath(
		session.Config.PointCount,
		session.Config.PressureMode,
	)

	if templatePath != "" {
		if _, err := os.Stat(templatePath); err == nil {
			return s.exportWithTemplate(templatePath, outputPath, standardValues, channels, unit, session)
		}
	}

	// 无模板，创建默认工作簿
	return s.exportFallback(outputPath, standardValues, channels, unit)
}

// exportWithTemplate 使用模板文件导出报告。
func (s *Service) exportWithTemplate(templatePath, outputPath string, standardValues []float64, channels [][]float64, unit string, session *calibration.CalibrationSession) error {
	f, err := LoadTemplate(templatePath)
	if err != nil {
		return fmt.Errorf("load template: %w", err)
	}
	defer f.Close()

	blocks, err := FindChannelBlocks(f)
	if err != nil {
		return fmt.Errorf("find channel blocks: %w", err)
	}

	for i, block := range blocks {
		if i >= len(channels) {
			break
		}

		// B 列填标准值（仅第一个块填充）
		if i == 0 {
			if err := FillStandardValues(f, block, "B", standardValues, unit); err != nil {
				return fmt.Errorf("fill standard values: %w", err)
			}
		}

		// C 列填测量值
		header := fmt.Sprintf("测量值-块%d", i+1)
		if err := FillMeasureData(f, block, "C", header, channels[i]); err != nil {
			return fmt.Errorf("fill measure data block %d: %w", i+1, err)
		}

		// 回程模式：D 列填回程数据
		if session.Config.PressureMode == "roundTrip" {
			backwardData := collectBackwardData(session, i)
			if len(backwardData) > 0 {
				if err := FillRoundTripData(f, block, "D", channels[i], backwardData); err != nil {
					return fmt.Errorf("fill round-trip data block %d: %w", i+1, err)
				}
			}
		}
	}

	if err := f.SaveAs(outputPath); err != nil {
		return fmt.Errorf("save report: %w", err)
	}

	return nil
}

// exportFallback 创建无模板的默认报告。
func (s *Service) exportFallback(outputPath string, standardValues []float64, channels [][]float64, unit string) error {
	f := CreateFallbackWorkbook(standardValues, channels, unit)
	defer f.Close()

	if err := f.SaveAs(outputPath); err != nil {
		return fmt.Errorf("save fallback report: %w", err)
	}

	return nil
}

// GetTemplates 返回模板目录中可用的模板列表与元信息。
func (s *Service) GetTemplates() ([]ReportTemplate, error) {
	entries, err := os.ReadDir(s.templateDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read template dir: %w", err)
	}

	templates := make([]ReportTemplate, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		template, ok := parseTemplateFileName(entry.Name())
		if !ok {
			continue
		}
		template.Path = filepath.Join(s.templateDir, entry.Name())
		templates = append(templates, template)
	}

	sort.Slice(templates, func(i, j int) bool {
		if templates[i].PointCount != templates[j].PointCount {
			return templates[i].PointCount < templates[j].PointCount
		}
		if templates[i].Mode != templates[j].Mode {
			return templates[i].Mode < templates[j].Mode
		}
		return templates[i].Name < templates[j].Name
	})

	return templates, nil
}

// MatchTemplate 根据点数与模式匹配模板绝对路径。
func (s *Service) MatchTemplate(pointCount int, mode string) (string, error) {
	filename, err := SelectTemplate(pointCount, mode)
	if err != nil {
		return "", err
	}

	fullPath := filepath.Join(s.templateDir, filename)
	if _, err := os.Stat(fullPath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("template not found: %s", filename)
		}
		return "", fmt.Errorf("check template: %w", err)
	}

	return fullPath, nil
}

func parseTemplateFileName(filename string) (ReportTemplate, bool) {
	ext := filepath.Ext(filename)
	if !strings.EqualFold(ext, ".xlsx") {
		return ReportTemplate{}, false
	}

	base := strings.TrimSuffix(filename, ext)
	if len(base) < 2 {
		return ReportTemplate{}, false
	}

	suffix := strings.ToLower(base[len(base)-1:])
	pointPart := base[:len(base)-1]
	pointCount, err := strconv.Atoi(pointPart)
	if err != nil || pointCount <= 0 {
		return ReportTemplate{}, false
	}

	mode := ""
	switch suffix {
	case "s":
		mode = "single"
	case "m":
		mode = "roundTrip"
	default:
		return ReportTemplate{}, false
	}

	return ReportTemplate{
		Name:       base,
		PointCount: pointCount,
		Mode:       mode,
	}, true
}

// collectChannelData 从会话压力点中按通道提取采集数据。
func collectChannelData(session *calibration.CalibrationSession) [][]float64 {
	if len(session.Points) == 0 {
		return nil
	}

	// 确定通道数
	numChannels := 0
	for _, p := range session.Points {
		if len(p.CollectedData) > numChannels {
			numChannels = len(p.CollectedData)
			break
		}
	}
	if numChannels == 0 {
		return nil
	}

	// 按通道聚合正程数据
	channels := make([][]float64, numChannels)
	for i := range channels {
		channels[i] = make([]float64, 0)
	}

	for _, p := range session.Points {
		if p.Direction == "backward" {
			continue
		}
		for ch := 0; ch < numChannels && ch < len(p.CollectedData); ch++ {
			channels[ch] = append(channels[ch], p.CollectedData[ch])
		}
	}

	return channels
}

// collectBackwardData 从会话压力点中提取指定通道的回程数据。
func collectBackwardData(session *calibration.CalibrationSession, channelIdx int) []float64 {
	var data []float64
	for _, p := range session.Points {
		if p.Direction != "backward" {
			continue
		}
		if channelIdx < len(p.CollectedData) {
			data = append(data, p.CollectedData[channelIdx])
		}
	}
	return data
}
