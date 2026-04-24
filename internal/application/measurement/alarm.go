package measurement

import (
	"fmt"
	"math"
)

type AlarmConfig struct {
	Enabled         bool    `json:"enabled"`
	EnabledChannels []int   `json:"enabledChannels"`
	ConfirmOnAlarm  bool    `json:"confirmOnAlarm"`
	SoundEnabled    bool    `json:"soundEnabled"`
	Threshold       float64 `json:"threshold"`
	IsRelative      bool    `json:"isRelative"`
}

type Alarm struct {
	PointID          string  `json:"pointId"`
	TargetPressure   float64 `json:"targetPressure"`
	ActualPressure   float64 `json:"actualPressure"`
	Threshold        float64 `json:"threshold"`
	IsRelative       bool    `json:"isRelative"`
	MaxDeviation     float64 `json:"maxDeviation"`
	OverLimitChannels []int  `json:"overLimitChannels"`
}

func (s *Service) SetAlarmConfig(cfg AlarmConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.alarmConfig = cfg
}

func (s *Service) GetAlarmConfig() AlarmConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.alarmConfig
}

func (s *Service) CheckAlarm(point Point) (*Alarm, error) {
	s.mu.Lock()
	cfg := s.alarmConfig
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

	var overLimit []int
	maxDev := 0.0

	for _, ch := range enabledCh {
		if ch < 1 || ch > len(point.CollectedData) {
			continue
		}
		collectedVal := point.CollectedData[ch-1]
		dev := math.Abs(collectedVal - point.TargetPressure)

		var exceeds bool
		if cfg.IsRelative && point.TargetPressure != 0 {
			exceeds = dev/point.TargetPressure > cfg.Threshold
		} else {
			exceeds = dev > cfg.Threshold
		}

		if exceeds {
			overLimit = append(overLimit, ch)
			if dev > maxDev {
				maxDev = dev
			}
		}
	}

	if len(overLimit) > 0 {
		alarm := &Alarm{
			PointID:           point.ID,
			TargetPressure:    point.TargetPressure,
			ActualPressure:    point.ActualPressure,
			Threshold:         cfg.Threshold,
			IsRelative:        cfg.IsRelative,
			MaxDeviation:      maxDev,
			OverLimitChannels: overLimit,
		}

		s.mu.Lock()
		s.alarmPending = true
		s.currentAlarm = alarm
		s.mu.Unlock()

		s.publish("measurement.alarm.triggered", alarm)

		return alarm, nil
	}

	return nil, nil
}

func (s *Service) IsAlarmPending() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.alarmPending
}

func (s *Service) ResolveAlarm(decision string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.alarmPending {
		return fmt.Errorf("no alarm pending")
	}

	s.publish("measurement.alarm.resolved", map[string]string{
		"decision": decision,
		"pointId":  s.currentAlarm.PointID,
	})

	s.alarmPending = false
	s.currentAlarm = nil

	return nil
}