package report

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ChannelBlock 表示模板中一个通道数据块的定位信息。
type ChannelBlock struct {
	Sheet     string
	HeaderRow int
	DataStart int
	DataEnd   int
}

// LoadTemplate 加载 Excel 模板文件并返回文件对象。
func LoadTemplate(path string) (*excelize.File, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("template file not found: %s", path)
	}
	return excelize.OpenFile(path)
}

// FindChannelBlocks 扫描模板所有工作表的 A 列，查找包含"通道"或"Channel"关键字的单元格，
// 将连续的数据行合并为一个 ChannelBlock。
func FindChannelBlocks(f *excelize.File) ([]ChannelBlock, error) {
	var blocks []ChannelBlock

	for _, sheet := range f.GetSheetList() {
		rows, err := f.GetCols(sheet)
		if err != nil || len(rows) == 0 {
			continue
		}

		colA := rows[0]
		var currentBlock *ChannelBlock

		for rowIdx, cell := range colA {
			rowNum := rowIdx + 1
			text := strings.TrimSpace(cell)

			if strings.Contains(text, "通道") || strings.Contains(text, "Channel") {
				if currentBlock != nil {
					blocks = append(blocks, *currentBlock)
				}
				currentBlock = &ChannelBlock{
					Sheet:     sheet,
					HeaderRow: rowNum,
					DataStart: rowNum + 1,
				}
			} else if currentBlock != nil && text == "" {
				currentBlock.DataEnd = rowNum - 1
				blocks = append(blocks, *currentBlock)
				currentBlock = nil
			}
		}

		if currentBlock != nil {
			currentBlock.DataEnd = len(colA)
			blocks = append(blocks, *currentBlock)
		}
	}

	return blocks, nil
}

// FillStandardValues 将标准压力值填入指定通道块的指定列。
func FillStandardValues(f *excelize.File, block ChannelBlock, col string, standardValues []float64, unit string) error {
	axis := fmt.Sprintf("%s%d", col, block.HeaderRow)
	if err := f.SetCellValue(block.Sheet, axis, fmt.Sprintf("标准值(%s)", unit)); err != nil {
		return err
	}

	for i, val := range standardValues {
		cell := fmt.Sprintf("%s%d", col, block.DataStart+i)
		rounded := math.Round(val*100) / 100
		if err := f.SetCellValue(block.Sheet, cell, rounded); err != nil {
			return err
		}
	}
	return nil
}

// FillMeasureData 将采集到的测量数据填入指定通道块的指定列。
func FillMeasureData(f *excelize.File, block ChannelBlock, col string, header string, data []float64) error {
	axis := fmt.Sprintf("%s%d", col, block.HeaderRow)
	if err := f.SetCellValue(block.Sheet, axis, header); err != nil {
		return err
	}

	for i, val := range data {
		cell := fmt.Sprintf("%s%d", col, block.DataStart+i)
		rounded := math.Round(val*1e6) / 1e6
		if err := f.SetCellValue(block.Sheet, cell, rounded); err != nil {
			return err
		}
	}
	return nil
}

// FillRoundTripData 将回程测量数据填入指定列（正程+回程）。
func FillRoundTripData(f *excelize.File, block ChannelBlock, col string, forwardData, backwardData []float64) error {
	axis := fmt.Sprintf("%s%d", col, block.HeaderRow)
	if err := f.SetCellValue(block.Sheet, axis, "回程测量值"); err != nil {
		return err
	}

	allData := append(forwardData, backwardData...)
	for i, val := range allData {
		row := block.DataStart + i
		if row > block.DataEnd {
			break
		}
		cell := fmt.Sprintf("%s%d", col, row)
		rounded := math.Round(val*1e6) / 1e6
		if err := f.SetCellValue(block.Sheet, cell, rounded); err != nil {
			return err
		}
	}
	return nil
}

// CreateFallbackWorkbook 当无模板文件时创建默认工作簿。
func CreateFallbackWorkbook(points []float64, channels [][]float64, unit string) *excelize.File {
	f := excelize.NewFile()
	sheet := "校准数据"
	f.SetSheetName("Sheet1", sheet)

	f.SetCellValue(sheet, "A1", "序号")
	f.SetCellValue(sheet, "B1", fmt.Sprintf("标准值(%s)", unit))

	for chIdx := range channels {
		col := fmt.Sprintf("%c", 'C'+chIdx)
		f.SetCellValue(sheet, fmt.Sprintf("%s1", col), fmt.Sprintf("通道%d", chIdx+1))
	}

	for i, p := range points {
		row := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), i+1)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), math.Round(p*100)/100)
		for chIdx, chData := range channels {
			if i < len(chData) {
				col := fmt.Sprintf("%c", 'C'+chIdx)
				f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, row), math.Round(chData[i]*1e6)/1e6)
			}
		}
	}

	return f
}

// ResolveUnit 按优先级解析压力单位：deviceUnit > cachedUnit > dataUnit > defaultUnit。
func ResolveUnit(deviceUnit, cachedUnit, dataUnit, defaultUnit string) string {
	if deviceUnit != "" {
		return deviceUnit
	}
	if cachedUnit != "" {
		return cachedUnit
	}
	if dataUnit != "" {
		return dataUnit
	}
	if defaultUnit != "" {
		return defaultUnit
	}
	return "kPa"
}
