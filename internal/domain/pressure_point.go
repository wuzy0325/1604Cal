package domain

import "math"

// 压力点状态常量，标定和计量模块共用。
const (
	PointStatusPending     = "pending"
	PointStatusPressurizing = "pressurizing"
	PointStatusStabilizing  = "stabilizing"
	PointStatusCollecting   = "collecting"
	PointStatusCompleted    = "completed"
	PointStatusError        = "error"
	PointStatusSkipped      = "skipped"
)

// AverageSamples 对多次采样的逐通道数据取平均，标定和计量模块共用。
// 各次采样的通道数可不完全一致——缺失通道不计入对应索引的平均值。
func AverageSamples(samples [][]float64) []float64 {
	if len(samples) == 0 {
		return nil
	}
	width := len(samples[0])
	result := make([]float64, width)
	counts := make([]int, width)
	for _, sample := range samples {
		for i := 0; i < len(sample) && i < width; i++ {
			result[i] += sample[i]
			counts[i]++
		}
	}
	for i := range result {
		if counts[i] > 0 {
			result[i] /= float64(counts[i])
		}
	}
	return result
}

// RoundToPrecision 按指定小数位精度对值进行四舍五入。
func RoundToPrecision(value float64, precision int) float64 {
	if precision <= 0 {
		return math.Round(value)
	}
	factor := math.Pow10(precision)
	return math.Round(value*factor) / factor
}

// EquidistantPoints 根据压力范围和点数生成等距压力点，含正程和可选回程。
// minPressure 和 maxPressure 的顺序会被自动修正。
func EquidistantPoints(minPressure, maxPressure float64, pointCount, precision int, roundTrip bool) []PressurePoint {
	if maxPressure < minPressure {
		minPressure, maxPressure = maxPressure, minPressure
	}
	if pointCount < 2 {
		pointCount = 2
	}

	step := (maxPressure - minPressure) / float64(pointCount-1)
	forward := make([]PressurePoint, pointCount)
	for i := 0; i < pointCount; i++ {
		forward[i] = PressurePoint{
			Index:          i + 1,
			TargetPressure: RoundToPrecision(minPressure+step*float64(i), precision),
			Direction:      "forward",
			Status:         PointStatusPending,
		}
	}

	if !roundTrip {
		return forward
	}

	backward := make([]PressurePoint, pointCount-1)
	for i := 0; i < pointCount-1; i++ {
		backward[i] = PressurePoint{
			Index:          pointCount + i + 1,
			TargetPressure: RoundToPrecision(maxPressure-step*float64(i), precision),
			Direction:      "backward",
			Status:         PointStatusPending,
		}
	}

	return append(forward, backward...)
}

// PressurePoint 表示一个压力测试点及其采集状态。
// 标定和计量模块共用此类型。
type PressurePoint struct {
	ID             string    `json:"id"`
	Index          int       `json:"index"`
	TargetPressure float64   `json:"targetPressure"`
	Direction      string    `json:"direction"` // forward | backward
	Status         string    `json:"status"`
	ActualPressure *float64  `json:"actualPressure,omitempty"`
	CollectedData  []float64 `json:"collectedData,omitempty"`
	CollectTime    string    `json:"collectTime,omitempty"`
	ErrorMessage   string    `json:"errorMessage,omitempty"`
}
