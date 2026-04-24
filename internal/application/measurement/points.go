package measurement

import (
	"fmt"
	"math"
	"strconv"

	apperrors "cal1604/internal/errors"
)

// Config 表示计量模块独立配置。
// 当前阶段先由 measurement 模块自己持有参数与点位计划，
// 后续自动/手动采集流程在此基础上继续扩展。
type Config struct {
	Channels       []int     `json:"channels"`
	MinPressure    float64   `json:"minPressure"`
	MaxPressure    float64   `json:"maxPressure"`
	PointCount     int       `json:"pointCount"`
	Precision      int       `json:"precision"`
	AverageCount   int       `json:"averageCount"`
	PrecisionLevel float64   `json:"precisionLevel"`
	StableWaitMs   int       `json:"stableWaitMs"`
	PressureMode   string    `json:"pressureMode"`
	ControlMode    string    `json:"controlMode"`
	CustomPoints   []float64 `json:"customPoints"`
}

// Point 表示计量模块测点计划中的单个压力点。
type Point struct {
	ID             string    `json:"id"`
	Index          int       `json:"index"`
	TargetPressure float64   `json:"targetPressure"`
	Direction      string    `json:"direction"`
	Status         string    `json:"status"`
	ActualPressure float64   `json:"actualPressure,omitempty"`
	CollectedData  []float64 `json:"collectedData,omitempty"`
	CollectTime    string    `json:"collectTime,omitempty"`
	ErrorMessage   string    `json:"errorMessage,omitempty"`
}

// SetConfig 更新计量模块当前配置。
func (s *Service) SetConfig(config Config) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = config
	s.points = nil
	s.session = nil
	if len(config.Channels) > 0 {
		s.channels = append([]int(nil), config.Channels...)
	}
}

// GetConfig 返回当前计量配置快照。
func (s *Service) GetConfig() Config {
	s.mu.Lock()
	defer s.mu.Unlock()
	config := s.config
	if len(config.Channels) > 0 {
		config.Channels = append([]int(nil), config.Channels...)
	}
	return config
}

// GeneratePressurePoints 根据 measurement 自己的配置生成测点计划。
func (s *Service) GeneratePressurePoints() ([]Point, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	points, err := generatePointsFromConfig(s.config)
	if err != nil {
		return nil, err
	}

	s.points = points
	result := make([]Point, len(points))
	copy(result, points)
	return result, nil
}

// GetPoints 返回当前计量测点快照。
func (s *Service) GetPoints() []Point {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]Point, len(s.points))
	copy(result, s.points)
	return result
}

func generatePointsFromConfig(config Config) ([]Point, error) {
	// 如果提供了自定义压力点，优先使用。
	if len(config.CustomPoints) > 0 {
		return generatePointsFromCustom(config)
	}

	if config.PointCount < 2 {
		return nil, fmt.Errorf("%w: point count must be at least 2, got %d", apperrors.ErrInvalidArgument, config.PointCount)
	}

	minPressure := config.MinPressure
	maxPressure := config.MaxPressure
	if maxPressure < minPressure {
		minPressure, maxPressure = maxPressure, minPressure
	}

	precision := config.Precision
	if precision < 0 {
		precision = 0
	}

	step := 0.0
	if config.PointCount > 1 {
		step = (maxPressure - minPressure) / float64(config.PointCount-1)
	}

	points := make([]Point, 0, config.PointCount*2-1)
	for i := 0; i < config.PointCount; i++ {
		pressure := roundToPrecision(minPressure+step*float64(i), precision)
		points = append(points, Point{
			ID:             "measurement-point-" + strconv.Itoa(i+1),
			Index:          i + 1,
			TargetPressure: pressure,
			Direction:      "forward",
			Status:         "pending",
		})
	}

	if config.PressureMode != "roundTrip" {
		return points, nil
	}

	for i := config.PointCount - 2; i >= 0; i-- {
		pressure := roundToPrecision(minPressure+step*float64(i), precision)
		points = append(points, Point{
			ID:             "measurement-point-" + strconv.Itoa(len(points)+1),
			Index:          len(points) + 1,
			TargetPressure: pressure,
			Direction:      "backward",
			Status:         "pending",
		})
	}

	return points, nil
}

// generatePointsFromCustom 根据用户自定义压力值生成测点计划。
func generatePointsFromCustom(config Config) ([]Point, error) {
	if len(config.CustomPoints) < 1 {
		return nil, fmt.Errorf("%w: custom points must not be empty", apperrors.ErrInvalidArgument)
	}

	precision := config.Precision
	if precision < 0 {
		precision = 0
	}

	points := make([]Point, 0, len(config.CustomPoints)*2-1)
	for i, pressure := range config.CustomPoints {
		rounded := roundToPrecision(pressure, precision)
		points = append(points, Point{
			ID:             "measurement-point-" + strconv.Itoa(i+1),
			Index:          i + 1,
			TargetPressure: rounded,
			Direction:      "forward",
			Status:         "pending",
		})
	}

	if config.PressureMode != "roundTrip" {
		return points, nil
	}

	for i := len(config.CustomPoints) - 2; i >= 0; i-- {
		rounded := roundToPrecision(config.CustomPoints[i], precision)
		points = append(points, Point{
			ID:             "measurement-point-" + strconv.Itoa(len(points)+1),
			Index:          len(points) + 1,
			TargetPressure: rounded,
			Direction:      "backward",
			Status:         "pending",
		})
	}

	return points, nil
}

func roundToPrecision(value float64, precision int) float64 {
	if precision <= 0 {
		return math.Round(value)
	}
	factor := math.Pow10(precision)
	return math.Round(value*factor) / factor
}
