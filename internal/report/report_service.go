package report

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"math"
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
// embedPrefix 为 embedFS 内的模板目录前缀，默认 "templates/reports"。
// 优先使用外部目录，不存在时回退到 embed.FS。
func NewService(templateDir string, embedFS ...fs.FS) *Service {
	s := &Service{templateDir: templateDir}
	if len(embedFS) > 0 && embedFS[0] != nil {
		s.embedTemplateProvider = NewEmbedTemplateProvider(embedFS[0], "templates/reports")
	}
	return s
}

// NewServiceWithPrefix 创建报告服务并指定 embed 前缀。
func NewServiceWithPrefix(templateDir string, embedFS fs.FS, embedPrefix string) *Service {
	s := &Service{templateDir: templateDir}
	if embedFS != nil {
		s.embedTemplateProvider = NewEmbedTemplateProvider(embedFS, embedPrefix)
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
	templatePath, err := s.ResolveTemplatePath(
		session.Config.PointCount,
		string(session.Config.PressureMode),
	)
	if err == nil && templatePath != "" {
		return s.exportWithTemplate(templatePath, outputPath, standardValues, channels, unit, session)
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
			sortTemplates(templates)
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

	sortTemplates(templates)
	return templates, nil
}

// sortTemplates 按点数、模式、名称排序模板列表。
func sortTemplates(templates []ReportTemplate) {
	sort.Slice(templates, func(i, j int) bool {
		if templates[i].PointCount != templates[j].PointCount {
			return templates[i].PointCount < templates[j].PointCount
		}
		if templates[i].Mode != templates[j].Mode {
			return templates[i].Mode < templates[j].Mode
		}
		return templates[i].Name < templates[j].Name
	})
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
	standardValues := collectMeasurementStandardValues(points, "forward")
	// 按通道聚合正程采集数据（从平铺数据计算每通道平均值）
	forwardChannels := collectMeasurementChannelData(points, numChannels, config.AverageCount, "forward")
	// 回程模式下额外按 targetPressure 索引聚合回程数据，避免回程点缺失时与正程错位；
	// 单程模式 backwardByTarget 为空。
	var backwardByTarget []map[float64]float64
	if config.PressureMode == domain.PressureModeRoundTrip {
		backwardByTarget = collectMeasurementChannelByTarget(points, numChannels, config.AverageCount, "backward")
	}
	unit := "kPa"

	// 尝试加载模板
	templatePath, err := s.ResolveTemplatePath(
		config.PointCount,
		string(config.PressureMode),
	)
	if err == nil && templatePath != "" {
		return s.exportMeasurementWithTemplate(ctx, templatePath, outputPath, standardValues, forwardChannels, backwardByTarget, unit, points, config)
	}

	// 无模板，创建默认工作簿
	return s.exportMeasurementFallback(outputPath, standardValues, forwardChannels, backwardByTarget, unit, points, config)
}

// exportMeasurementWithTemplate 使用模板文件导出计量报告。
// 计量模板列映射：
//   单程模板 *s.xlsx：A=标准压力，B=设备显示值，C=示值误差(公式)，D=不确定度
//   回程模板 *m.xlsx：A=标准压力，B=正程显示值(Forward stroke)，C=回程显示值(Return stroke)，D=示值误差(公式)，E=回差(公式)
// backwardByTarget 按 (通道, 标准压力) 索引，确保 C 列严格按 A 列标准值对齐，
// 即便部分回程点未完成也不会发生静默错位。
func (s *Service) exportMeasurementWithTemplate(ctx context.Context, templatePath, outputPath string, standardValues []float64, forwardChannels [][]float64, backwardByTarget []map[float64]float64, unit string, points []domain.PressurePoint, config domain.WorkflowConfig) error {
	f, err := LoadTemplate(templatePath)
	if err != nil {
		return fmt.Errorf("%w: load template: %v", apperrors.ErrReportExport, err)
	}
	defer f.Close()

	blocks, err := FindChannelBlocks(f)
	if err != nil {
		return fmt.Errorf("%w: find channel blocks: %v", apperrors.ErrReportExport, err)
	}

	isRoundTrip := config.PressureMode == domain.PressureModeRoundTrip

	for i, block := range blocks {
		if i >= len(forwardChannels) {
			break
		}

		// A 列填标准压力
		for j, val := range standardValues {
			cell := fmt.Sprintf("A%d", block.DataStart+j)
			rounded := math.Round(val*100) / 100
			if err := f.SetCellValue(block.Sheet, cell, rounded); err != nil {
				return fmt.Errorf("%w: fill standard values block %d row %d: %v", apperrors.ErrReportExport, i+1, j+1, err)
			}
		}

		// B 列填正程显示值（单程模式即唯一显示值）
		for j, val := range forwardChannels[i] {
			cell := fmt.Sprintf("B%d", block.DataStart+j)
			rounded := math.Round(val*1e6) / 1e6
			if err := f.SetCellValue(block.Sheet, cell, rounded); err != nil {
				return fmt.Errorf("%w: fill measure data block %d row %d: %v", apperrors.ErrReportExport, i+1, j+1, err)
			}
		}

		// 回程模式：C 列按 A 列标准压力精确匹配回程显示值（Return stroke），
		// 缺失的回程点保留模板初值，绝不写错行。
		if isRoundTrip && i < len(backwardByTarget) {
			lookup := backwardByTarget[i]
			for j, std := range standardValues {
				val, ok := lookup[std]
				if !ok {
					continue
				}
				cell := fmt.Sprintf("C%d", block.DataStart+j)
				rounded := math.Round(val*1e6) / 1e6
				if err := f.SetCellValue(block.Sheet, cell, rounded); err != nil {
					return fmt.Errorf("%w: fill return stroke block %d row %d: %v", apperrors.ErrReportExport, i+1, j+1, err)
				}
			}
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
// 回程模式下若 backwardByTarget 非空，会同时写入"回程值"列。
func (s *Service) exportMeasurementFallback(outputPath string, standardValues []float64, forwardChannels [][]float64, backwardByTarget []map[float64]float64, unit string, points []domain.PressurePoint, config domain.WorkflowConfig) error {
	f := CreateMeasurementFallbackWorkbook(standardValues, forwardChannels, backwardByTarget, unit, points, config)
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

	// 从 points 提取首条采集时间（如有），转换为可读格式
	startTime := nowStr
	for _, p := range points {
		if p.CollectTime != "" {
			if t, err := time.Parse(time.RFC3339, p.CollectTime); err == nil {
				startTime = t.Format("2006-01-02 15:04:05")
			} else {
				startTime = p.CollectTime
			}
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

			// 匹配"Min(Range)"标签→右侧单元格填充最小量程
			if strings.Contains(text, "Min(Range)") || strings.Contains(text, "Min（Range）") {
				rightCell := cellName(col+1, row)
				f.SetCellValue(sheet, rightCell, config.MinPressure)
				continue
			}

			// 匹配"Max(Range)"标签→右侧单元格填充最大量程
			if strings.Contains(text, "Max(Range)") || strings.Contains(text, "Max（Range）") {
				rightCell := cellName(col+1, row)
				f.SetCellValue(sheet, rightCell, config.MaxPressure)
				continue
			}

			// 匹配"Accuracy"标签→右侧单元格填充准确度等级（按百分数显示，如 0.02 表示 0.02%）
			if strings.Contains(text, "Accuracy") || strings.Contains(text, "准确度") {
				rightCell := cellName(col+1, row)
				f.SetCellValue(sheet, rightCell, fmt.Sprintf("%.2f", config.PrecisionLevel*100))
				continue
			}

			// 匹配"Equipment Number"或"设备编号"标签→右侧单元格填充设备编号
			if strings.Contains(text, "Equipment Number") || strings.Contains(text, "设备编号") {
				rightCell := cellName(col+1, row)
				if config.DeviceNumber != "" {
					f.SetCellValue(sheet, rightCell, config.DeviceNumber)
				}
				continue
			}
		}
	}
}

// matchDirection 按指定方向过滤压力点。
// direction="forward" 接受 Direction 为空或 "forward"；direction="backward" 仅接受 "backward"。
func matchDirection(pointDirection, direction string) bool {
	if direction == "backward" {
		return pointDirection == "backward"
	}
	// 默认按正程：兼容旧数据 Direction 为空的情况
	return pointDirection == "" || pointDirection == "forward"
}

// collectMeasurementStandardValues 从计量压力点中提取指定方向已完成的标准值。
// 过滤条件与 collectMeasurementChannelData 保持一致，确保标准值和通道数据一一对应。
func collectMeasurementStandardValues(points []domain.PressurePoint, direction string) []float64 {
	values := make([]float64, 0, len(points))
	for _, p := range points {
		if !matchDirection(p.Direction, direction) || p.Status != "completed" || len(p.CollectedData) == 0 {
			continue
		}
		values = append(values, p.TargetPressure)
	}
	return values
}

// collectMeasurementChannelByTarget 按 (通道, 标准压力) 聚合指定方向的平均值，
// 返回 channels[ch][targetPressure] = avg。键直接使用 float64 的 TargetPressure
// （等距生成时由 RoundToPrecision 量化，存在精确相等保证）。
// 跳过未完成、CollectedData 为空、以及无法推断 samplesPerChannel 的异常点。
func collectMeasurementChannelByTarget(points []domain.PressurePoint, numChannels, averageCount int, direction string) []map[float64]float64 {
	result := make([]map[float64]float64, numChannels)
	for i := range result {
		result[i] = make(map[float64]float64)
	}

	for _, p := range points {
		if !matchDirection(p.Direction, direction) || p.Status != "completed" {
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
			log.Printf("report: skip measurement point index=%d direction=%s: cannot infer samplesPerChannel from len(CollectedData)=%d, numChannels=%d",
				p.Index, p.Direction, len(p.CollectedData), numChannels)
			continue
		}

		for ch := 0; ch < numChannels; ch++ {
			sum := 0.0
			count := 0
			for sIdx := 0; sIdx < samplesPerChannel; sIdx++ {
				idx := sIdx*numChannels + ch
				if idx < len(p.CollectedData) {
					sum += p.CollectedData[idx]
					count++
				}
			}
			if count == 0 {
				continue
			}
			result[ch][p.TargetPressure] = sum / float64(count)
		}
	}

	return result
}

// collectMeasurementChannelData 从平铺的 CollectedData 中按通道聚合指定方向的平均值。
// 计量模块的 CollectedData 是平铺格式：sample0_ch0, sample0_ch1, ..., sampleN_ch0, sampleN_ch1。
// direction 为 "forward" 时仅聚合正程点，为 "backward" 时仅聚合回程点。
// 返回顺序与 points 中匹配方向点的出现顺序一致；如需按标准压力查找回程数据，
// 请改用 collectMeasurementChannelByTarget。
func collectMeasurementChannelData(points []domain.PressurePoint, numChannels, averageCount int, direction string) [][]float64 {
	channels := make([][]float64, numChannels)
	for i := range channels {
		channels[i] = make([]float64, 0)
	}

	for _, p := range points {
		if !matchDirection(p.Direction, direction) || p.Status != "completed" {
			continue
		}
		if len(p.CollectedData) == 0 {
			continue
		}

		samplesPerChannel := averageCount
		if samplesPerChannel <= 0 {
			samplesPerChannel = len(p.CollectedData) / numChannels
		}
		// 异常数据点直接跳过，避免向通道写入 0 导致下游对齐错误。
		if samplesPerChannel <= 0 {
			log.Printf("report: skip measurement point index=%d direction=%s: cannot infer samplesPerChannel from len(CollectedData)=%d, numChannels=%d",
				p.Index, p.Direction, len(p.CollectedData), numChannels)
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
