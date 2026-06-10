package report

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"cal1604/internal/application/calibration"
	"cal1604/internal/domain"
	apperrors "cal1604/internal/errors"

	"github.com/xuri/excelize/v2"
)

// Service 封装报告模板路径拼装逻辑与报告导出。
type Service struct {
	templateDir         string
	embedTemplateProvider *EmbedTemplateProvider
}

// ReportTemplate 描述一个可用报告模板。
type ReportTemplate struct {
	Name       string `json:"name"`
	PointCount int    `json:"pointCount"`
	Mode       string `json:"mode"`
	Path       string `json:"path"`
}

// NewService 创建报告服务。
// templateDir 为外部模板目录（可选），embedFS 为嵌入模板文件系统（可选）。
// 优先使用外部目录，不存在时回退到 embed.FS。
func NewService(templateDir string, embedFS ...fs.FS) *Service {
	s := &Service{templateDir: templateDir}
	if len(embedFS) > 0 && embedFS[0] != nil {
		s.embedTemplateProvider = NewEmbedTemplateProvider(embedFS[0], "templates/reports")
	}
	return s
}

// SetEmbedTemplateProvider 设置嵌入模板提供者（用于运行时动态注入）。
func (s *Service) SetEmbedTemplateProvider(provider *EmbedTemplateProvider) {
	s.embedTemplateProvider = provider
}

// CleanupEmbedTemplates 清理嵌入模板解压的临时目录。
func (s *Service) CleanupEmbedTemplates() error {
	if s.embedTemplateProvider != nil {
		return s.embedTemplateProvider.Cleanup()
	}
	return nil
}

// ResolveTemplatePath 解析模板绝对路径。
func (s *Service) ResolveTemplatePath(points int, mode string) (string, error) {
	return s.MatchTemplate(points, mode)
}

// ExportReport 根据 CalibrationSession 生成校准报告并保存到 outputPath。
// 优先使用模板文件填充数据，无模板时创建默认工作簿。
func (s *Service) ExportReport(ctx context.Context, session *calibration.CalibrationSession, outputPath string) error {
	if session == nil {
		return fmt.Errorf("%w: calibration session is nil", apperrors.ErrNoActiveSession)
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
		string(session.Config.PressureMode),
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
		return fmt.Errorf("%w: load template: %v", apperrors.ErrReportExport, err)
	}
	defer f.Close()

	blocks, err := FindChannelBlocks(f)
	if err != nil {
		return fmt.Errorf("%w: find channel blocks: %v", apperrors.ErrReportExport, err)
	}

	for i, block := range blocks {
		if i >= len(channels) {
			break
		}

		// B 列填标准值（仅第一个块填充）
		if i == 0 {
			if err := FillStandardValues(f, block, "B", standardValues, unit); err != nil {
				return fmt.Errorf("%w: fill standard values: %v", apperrors.ErrReportExport, err)
			}
		}

		// C 列填测量值
		header := fmt.Sprintf("测量值-块%d", i+1)
		if err := FillMeasureData(f, block, "C", header, channels[i]); err != nil {
			return fmt.Errorf("%w: fill measure data block %d: %v", apperrors.ErrReportExport, i+1, err)
		}

		// 回程模式：D 列填回程数据
		if session.Config.PressureMode == domain.PressureModeRoundTrip {
			backwardData := collectBackwardData(session, i)
			if len(backwardData) > 0 {
				if err := FillRoundTripData(f, block, "D", channels[i], backwardData); err != nil {
					return fmt.Errorf("%w: fill round-trip data block %d: %v", apperrors.ErrReportExport, i+1, err)
				}
			}
		}
	}

	if err := f.SaveAs(outputPath); err != nil {
		return fmt.Errorf("%w: save report: %v", apperrors.ErrReportExport, err)
	}

	return nil
}

// exportFallback 创建无模板的默认报告。
func (s *Service) exportFallback(outputPath string, standardValues []float64, channels [][]float64, unit string) error {
	f := CreateFallbackWorkbook(standardValues, channels, unit)
	defer f.Close()

	if err := f.SaveAs(outputPath); err != nil {
		return fmt.Errorf("%w: save fallback report: %v", apperrors.ErrReportExport, err)
	}

	return nil
}

// GetTemplates 返回模板目录中可用的模板列表与元信息。
// 优先扫描外部目录，不存在时从 embed.FS 获取。
func (s *Service) GetTemplates() ([]ReportTemplate, error) {
	var entries []os.DirEntry
	var dirPath string
	var err error

	if s.templateDir != "" {
		entries, err = os.ReadDir(s.templateDir)
		dirPath = s.templateDir
	}
	if s.templateDir == "" || (err != nil && os.IsNotExist(err)) {
		if s.embedTemplateProvider != nil {
			files, listErr := s.embedTemplateProvider.ListTemplates()
			if listErr != nil {
				return nil, fmt.Errorf("%w: list embed templates: %v", apperrors.ErrReportExport, listErr)
			}
			templates := make([]ReportTemplate, 0, len(files))
			for _, name := range files {
				template, ok := parseTemplateFileName(name)
				if !ok {
					continue
				}
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
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("%w: read template dir: %v", apperrors.ErrReportExport, err)
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
		template.Path = filepath.Join(dirPath, entry.Name())
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
// 优先检查外部目录，其次从 embed.FS 解压到临时目录后返回路径。
func (s *Service) MatchTemplate(pointCount int, mode string) (string, error) {
	filename, err := SelectTemplate(pointCount, mode)
	if err != nil {
		return "", err
	}

	// 优先使用外部模板目录
	if s.templateDir != "" {
		fullPath := filepath.Join(s.templateDir, filename)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, nil
		}
	}

	// 回退到 embed.FS
	if s.embedTemplateProvider != nil {
		return s.embedTemplateProvider.ResolvePath(filename)
	}

	return "", fmt.Errorf("%w: template not found: %s", apperrors.ErrReportExport, filename)
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
		mode = string(domain.PressureModeSingle)
	case "m":
		mode = string(domain.PressureModeRoundTrip)
	default:
		return ReportTemplate{}, false
	}

	return ReportTemplate{
		Name:       base,
		PointCount: pointCount,
		Mode:       mode,
	}, true
}

// ExportMeasurementReport 根据计量采集数据生成报告并保存到 outputPath。
func (s *Service) ExportMeasurementReport(ctx context.Context, points []domain.PressurePoint, config domain.WorkflowConfig, outputPath string) error {
	if len(points) == 0 {
		return fmt.Errorf("%w: no measurement points", apperrors.ErrNoActiveSession)
	}

	// 始终输出全部16通道
	numChannels := 16

	// 提取正程标准压力值
	standardValues := collectMeasurementStandardValues(points)
	// 按通道聚合采集数据（从平铺数据计算每通道平均值）
	channels := collectMeasurementChannelData(points, numChannels, config.AverageCount)
	unit := "kPa"

	// 尝试加载模板
	templatePath, _ := s.ResolveTemplatePath(
		config.PointCount,
		string(config.PressureMode),
	)

	if templatePath != "" {
		if _, err := os.Stat(templatePath); err == nil {
			return s.exportMeasurementWithTemplate(ctx, templatePath, outputPath, standardValues, channels, unit, points, config)
		}
	}

	// 无模板，创建默认工作簿
	return s.exportMeasurementFallback(outputPath, standardValues, channels, unit, points, config)
}

// exportMeasurementWithTemplate 使用模板文件导出计量报告。
func (s *Service) exportMeasurementWithTemplate(ctx context.Context, templatePath, outputPath string, standardValues []float64, channels [][]float64, unit string, points []domain.PressurePoint, config domain.WorkflowConfig) error {
	f, err := LoadTemplate(templatePath)
	if err != nil {
		return fmt.Errorf("%w: load template: %v", apperrors.ErrReportExport, err)
	}
	defer f.Close()

	blocks, err := FindChannelBlocks(f)
	if err != nil {
		return fmt.Errorf("%w: find channel blocks: %v", apperrors.ErrReportExport, err)
	}

	for i, block := range blocks {
		if i >= len(channels) {
			break
		}

		if i == 0 {
			if err := FillStandardValues(f, block, "B", standardValues, unit); err != nil {
				return fmt.Errorf("%w: fill standard values: %v", apperrors.ErrReportExport, err)
			}
		}

		header := fmt.Sprintf("通道%d", i+1)
		if err := FillMeasureData(f, block, "C", header, channels[i]); err != nil {
			return fmt.Errorf("%w: fill measure data block %d: %v", apperrors.ErrReportExport, i+1, err)
		}
	}

	// 填充元数据：单位、日期等
	fillMeasurementWorksheetMetadata(f, unit, points, config)

	if err := f.SaveAs(outputPath); err != nil {
		return fmt.Errorf("%w: save report: %v", apperrors.ErrReportExport, err)
	}

	return nil
}

// exportMeasurementFallback 创建无模板的计量报告。
func (s *Service) exportMeasurementFallback(outputPath string, standardValues []float64, channels [][]float64, unit string, points []domain.PressurePoint, config domain.WorkflowConfig) error {
	f := CreateMeasurementFallbackWorkbook(standardValues, channels, unit, points, config)
	defer f.Close()

	fillMeasurementWorksheetMetadata(f, unit, points, config)

	if err := f.SaveAs(outputPath); err != nil {
		return fmt.Errorf("%w: save fallback report: %v", apperrors.ErrReportExport, err)
	}

	return nil
}

// fillMeasurementWorksheetMetadata 扫描工作表填写元数据字段（日期、单位等）。
func fillMeasurementWorksheetMetadata(f *excelize.File, unit string, points []domain.PressurePoint, config domain.WorkflowConfig) {
	if len(f.GetSheetList()) == 0 {
		return
	}
	sheet := f.GetSheetList()[0]

	nowStr := time.Now().Format("2006-01-02 15:04:05")

	// 从 points 提取首条采集时间（如有）
	startTime := nowStr
	for _, p := range points {
		if p.CollectTime != "" {
			startTime = p.CollectTime
			break
		}
	}

	// 扫描前 50 行的前 12 列，匹配中文标签
	for row := 1; row <= 50; row++ {
		for col := 1; col <= 12; col++ {
			cell := cellName(col, row)
			text, _ := f.GetCellValue(sheet, cell)
			text = strings.TrimSpace(text)

			// 匹配"校准日期"或"日期"标签→右侧单元格填充日期
			if strings.Contains(text, "日期") || strings.Contains(strings.ToLower(text), "date") {
				rightCell := cellName(col+1, row)
				f.SetCellValue(sheet, rightCell, startTime)
				continue
			}

			// 匹配"单位"标签→右侧单元格填充单位
			if strings.Contains(text, "单位") || strings.Contains(strings.ToLower(text), "unit") {
				if col+1 <= 12 {
					rightCell := cellName(col+1, row)
					val, _ := f.GetCellValue(sheet, rightCell)
					if val == "" || val == "kPa" {
						f.SetCellValue(sheet, rightCell, unit)
					}
				}
				continue
			}
		}
	}
}

// collectMeasurementStandardValues 从计量压力点中提取正程标准值。
func collectMeasurementStandardValues(points []domain.PressurePoint) []float64 {
	values := make([]float64, 0, len(points))
	for _, p := range points {
		if p.Direction == "backward" {
			continue
		}
		values = append(values, p.TargetPressure)
	}
	return values
}

// collectMeasurementChannelData 从平铺的 CollectedData 中按通道聚合平均值。
// 计量模块的 CollectedData 是平铺格式：sample0_ch0, sample0_ch1, ..., sampleN_ch0, sampleN_ch1。
func collectMeasurementChannelData(points []domain.PressurePoint, numChannels, averageCount int) [][]float64 {
	channels := make([][]float64, numChannels)
	for i := range channels {
		channels[i] = make([]float64, 0)
	}

	for _, p := range points {
		if p.Direction == "backward" || p.Status != "completed" {
			continue
		}
		if len(p.CollectedData) == 0 {
			continue
		}

		samplesPerChannel := averageCount
		if samplesPerChannel <= 0 {
			samplesPerChannel = len(p.CollectedData) / numChannels
		}
		if samplesPerChannel <= 0 {
			for ch := 0; ch < numChannels; ch++ {
				channels[ch] = append(channels[ch], 0)
			}
			continue
		}

		for ch := 0; ch < numChannels; ch++ {
			sum := 0.0
			count := 0
			for s := 0; s < samplesPerChannel; s++ {
				idx := s*numChannels + ch
				if idx < len(p.CollectedData) {
					sum += p.CollectedData[idx]
					count++
				}
			}
			avg := 0.0
			if count > 0 {
				avg = sum / float64(count)
			}
			channels[ch] = append(channels[ch], avg)
		}
	}

	return channels
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
