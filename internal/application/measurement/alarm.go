package measurement

import (
	"fmt"
	"math"

	"cal1604/internal/domain"
	"cal1604/internal/events"
)

type Alarm struct {
	PointID           string  `json:"pointId"`
	TargetPressure    float64 `json:"targetPressure"`
	ActualPressure    float64 `json:"actualPressure"`
	Threshold         float64 `json:"threshold"`
	MaxDeviation      float64 `json:"maxDeviation"`
	OverLimitChannels []int   `json:"overLimitChannels"`
}

func (s *Service) SetAlarmConfig(cfg domain.AlarmConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.alarmConfig = cfg
}

func (s *Service) GetAlarmConfig() domain.AlarmConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.alarmConfig
}

func (s *Service) CheckAlarm(point domain.PressurePoint) (*Alarm, error) {
	s.mu.Lock()
	cfg := s.alarmConfig
	workCfg := s.config
	s.mu.Unlock()

	if !cfg.Enabled {
		return nil, nil
	}

	if len(point.CollectedData) == 0 {
		return nil, nil
	}

	enabledCh := cfg.EnabledChannels
	if len(enabledCh) == 0 {
		for i := range point.CollectedData {
			enabledCh = append(enabledCh, i+1)
		}
	}

	// 量程引用误差：允许偏差 = 量程 x 准确度等级。
	// 当量程为 0（如配置异常或固定单点）时降级为按目标值比例计算。
	span := math.Abs(workCfg.MaxPressure - workCfg.MinPressure)
	allowance := span * workCfg.PrecisionLevel
	if allowance < 1e-10 {
		allowance = math.Abs(point.TargetPressure) * workCfg.PrecisionLevel
	}

	var overLimit []int
	maxDev := 0.0

	for _, ch := range enabledCh {
		if ch < 1 || ch > len(point.CollectedData) {
			continue
		}
		collectedVal := point.CollectedData[ch-1]
		dev := math.Abs(collectedVal - point.TargetPressure)

		if dev > allowance {
			overLimit = append(overLimit, ch)
			if dev > maxDev {
				maxDev = dev
			}
		}
	}

	if len(overLimit) == 0 {
		return nil, nil
	}

	// maxDeviation 表示最大偏差占量程的比值（FS 百分比小数形式）。
	// 量程为 0 时退化为相对目标值的比例，目标值也为 0 时为 0。
	var maxDevRatio float64
	switch {
	case span > 1e-10:
		maxDevRatio = maxDev / span
	case point.TargetPressure != 0:
		maxDevRatio = maxDev / math.Abs(point.TargetPressure)
	}

	var actualPressure float64
	if point.ActualPressure != nil {
		actualPressure = *point.ActualPressure
	}

	alarm := &Alarm{
		PointID:           point.ID,
		TargetPressure:    point.TargetPressure,
		ActualPressure:    actualPressure,
		Threshold:         workCfg.PrecisionLevel,
		MaxDeviation:      maxDevRatio,
		OverLimitChannels: overLimit,
	}

	s.mu.Lock()
	s.alarmPending = true
	s.currentAlarm = alarm
	s.mu.Unlock()

	s.publish(events.EventMeasurementAlarmTriggered, alarm)

	return alarm, nil
}

func (s *Service) IsAlarmPending() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.alarmPending
}

func (s *Service) ResolveAlarm(decision string) error {
	s.mu.Lock()
	if !s.alarmPending {
		s.mu.Unlock()
		return fmt.Errorf("no alarm pending")
	}

	alarmCh := s.alarmCh
	s.mu.Unlock()

	if alarmCh != nil {
		select {
		case alarmCh <- decision:
		default:
		}
	}

	s.mu.Lock()
	s.alarmPending = false
	s.currentAlarm = nil
	s.mu.Unlock()

	s.publish(events.EventMeasurementAlarmResolved, map[string]string{
		"decision": decision,
	})

	return nil
}

func (s *Service) ResolveStabilityTimeout(decision string) {
	select {
	case s.stabilityTimeoutCh <- decision:
	default:
	}
}
