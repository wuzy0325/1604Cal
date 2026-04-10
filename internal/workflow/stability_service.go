package workflow

import (
	"math"
	"time"
)

// StabilityAccumulator 用于累计稳定时间。
type StabilityAccumulator struct {
	tolerance        float64
	requiredDuration time.Duration
	accumulated      time.Duration
}

// NewStabilityAccumulator 创建稳定累计器。
func NewStabilityAccumulator(tolerance float64, requiredDuration time.Duration) *StabilityAccumulator {
	return &StabilityAccumulator{
		tolerance:        tolerance,
		requiredDuration: requiredDuration,
	}
}

// AddSample 增加一次采样并返回是否达稳以及当前累计时长。
func (a *StabilityAccumulator) AddSample(target, actual float64, interval time.Duration) (bool, time.Duration) {
	if interval < 0 {
		interval = 0
	}

	deviation := math.Abs(actual - target)
	if deviation <= a.tolerance {
		a.accumulated += interval
	} else {
		a.accumulated = 0
	}

	return a.accumulated >= a.requiredDuration, a.accumulated
}
