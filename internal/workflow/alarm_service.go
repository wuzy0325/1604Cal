package workflow

import (
	"fmt"
	"math"
)

const (
	alarmDecisionContinue = "continue"
	alarmDecisionRetry    = "retry"
)

// AlarmResult 表示一次报警判定结果。
type AlarmResult struct {
	Triggered        bool
	Deviation        float64
	DeviationPercent float64
	Allowance        float64
}

// AlarmService 负责计算精度超限并校验处置动作。
type AlarmService struct{}

// NewAlarmService 创建报警规则服务。
func NewAlarmService() *AlarmService {
	return &AlarmService{}
}

// Evaluate 根据目标值、实测值和百分比阈值计算报警结果。
func (s *AlarmService) Evaluate(target, actual, levelPercent float64) AlarmResult {
	deviation := math.Abs(actual - target)
	allowance := math.Abs(target) * levelPercent / 100

	deviationPercent := 0.0
	if target != 0 {
		deviationPercent = deviation / math.Abs(target) * 100
	}

	return AlarmResult{
		Triggered:        deviation > allowance,
		Deviation:        deviation,
		DeviationPercent: deviationPercent,
		Allowance:        allowance,
	}
}

// ValidateDecision 校验报警后的用户决策动作是否合法。
func (s *AlarmService) ValidateDecision(decision string) error {
	if decision == alarmDecisionContinue || decision == alarmDecisionRetry {
		return nil
	}

	return fmt.Errorf("invalid alarm decision: %s", decision)
}
