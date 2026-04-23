package calibration

import (
	"context"
	"fmt"
	"math"
	"time"

	"cal1604/internal/domain"
	"cal1604/internal/infrastructure/driver"
)

// calibrationPointCollector 定义支持“按校准点采集”的驱动能力。
// 例如 WTN1604 在进入 C00 校准流程后，需使用 C01 命令采集单点数据。
type calibrationPointCollector interface {
	CollectCalibrationPoint(ctx context.Context, pointIndex int, targetPressure float64) ([]float64, error)
}

// Collect 从计量设备采集数据。
func (s *Service) Collect(ctx context.Context, pointIndex int) ([]float64, error) {
	s.mu.Lock()
	s.syncDriversFromSessionLocked()
	if s.measureDriver == nil {
		s.mu.Unlock()
		return nil, ErrMeasureDeviceNotSet
	}
	if pointIndex < 1 || pointIndex > len(s.pressurePoints) {
		s.mu.Unlock()
		return nil, fmt.Errorf("invalid point index: %d", pointIndex)
	}
	measureDriver := s.measureDriver
	targetPressure := s.pressurePoints[pointIndex-1].TargetPressure
	channels := s.config.Channels
	avgCount := s.config.AverageCount
	if avgCount < 1 {
		avgCount = 1
	}
	s.mu.Unlock()

	s.updatePointStatus(pointIndex, "collecting")

	// 状态迁移: -> collecting
	if err := s.sessionMachine.Transition(domain.SessionStateCollecting); err != nil {
		// 可能已经在 collecting
	}
	s.publishSessionState()

	var averaged []float64

	// WTN1604 等设备在校准模式下需要走“按点采集”命令，
	// 不能继续复用普通实时采集命令，否则会出现“设备已连接但采集失败”。
	if pointCollector, ok := measureDriver.(calibrationPointCollector); ok {
		// 校准点采集同样遵循“多次采样取平均”配置，
		// 仅将单次采样命令从普通采集切换为 C01 校准点采集。
		allSamples := make([][]float64, 0, avgCount)
		for i := 0; i < avgCount; i++ {
			data, err := pointCollector.CollectCalibrationPoint(ctx, pointIndex, targetPressure)
			if err != nil {
				s.markPointError(pointIndex, err.Error())
				return nil, fmt.Errorf("collect calibration point %d sample %d: %w", pointIndex, i+1, err)
			}
			allSamples = append(allSamples, data)
			time.Sleep(100 * time.Millisecond)
		}
		averaged = averageSamples(allSamples)
	} else {
		// 多次采集求平均（非校准点采集驱动的通用路径）
		allSamples := make([][]float64, 0, avgCount)
		for i := 0; i < avgCount; i++ {
			data, err := measureDriver.CollectData(ctx, channels)
			if err != nil {
				s.markPointError(pointIndex, err.Error())
				return nil, fmt.Errorf("collect sample %d: %w", i+1, err)
			}
			allSamples = append(allSamples, data)
			time.Sleep(100 * time.Millisecond)
		}

		// 计算平均值
		averaged = averageSamples(allSamples)
	}

	s.mu.Lock()
	s.pressurePoints[pointIndex-1].CollectedData = averaged
	s.currentPoint = pointIndex
	s.mu.Unlock()

	s.updatePointStatus(pointIndex, "completed")

	// 状态迁移: collecting -> point_done
	if err := s.sessionMachine.Transition(domain.SessionStatePointDone); err != nil {
		// 忽略
	}
	s.publishSessionState()

	s.publish("data.collected", map[string]any{
		"pointIndex": pointIndex,
		"channels":   channels,
		"data":       averaged,
	})

	return averaged, nil
}

// Fit 执行数据拟合。
func (s *Service) Fit(ctx context.Context) (*FittingResult, error) {
	s.mu.Lock()
	if s.measureDriver == nil {
		s.mu.Unlock()
		return nil, ErrMeasureDeviceNotSet
	}
	s.mu.Unlock()

	// 状态迁移: -> fitting
	if err := s.sessionMachine.Transition(domain.SessionStateFitting); err != nil {
		return nil, fmt.Errorf("transition to fitting: %w", err)
	}
	s.publishSessionState()

	// WTN1604 执行拟合
	wtn, ok := s.measureDriver.(*driver.WTN1604Driver)
	if !ok {
		// 如果不是 WTN1604，使用软件拟合
		return s.softwareFit(ctx)
	}

	if err := wtn.PerformFitting(ctx); err != nil {
		return nil, fmt.Errorf("perform fitting: %w", err)
	}

	if err := wtn.SaveCoefficients(ctx); err != nil {
		return nil, fmt.Errorf("save coefficients: %w", err)
	}

	// 状态迁移: fitting -> completed
	if err := s.sessionMachine.Transition(domain.SessionStateCompleted); err != nil {
		return nil, fmt.Errorf("transition to completed: %w", err)
	}
	s.publishSessionState()

	// WTN1604 拟合不返回系数细节，返回占位结果
	return &FittingResult{Points: len(s.pressurePoints)}, nil
}

// softwareFit 软件侧拟合（非 WTN1604 设备的备选方案）。
func (s *Service) softwareFit(ctx context.Context) (*FittingResult, error) {
	// 简单线性拟合：y = a*x + b
	// 使用最小二乘法
	s.mu.Lock()
	points := s.pressurePoints
	s.mu.Unlock()

	var sumX, sumY, sumXY, sumX2 float64
	n := 0
	for _, p := range points {
		if p.Status == "completed" && len(p.CollectedData) > 0 {
			x := p.TargetPressure
			y := p.CollectedData[0] // 使用第一个通道的数据
			sumX += x
			sumY += y
			sumXY += x * y
			sumX2 += x * x
			n++
		}
	}

	if n < 2 {
		return nil, fmt.Errorf("not enough data points for fitting: %d", n)
	}

	// y = a*x + b
	denom := float64(n)*sumX2 - sumX*sumX
	if math.Abs(denom) < 1e-10 {
		return nil, fmt.Errorf("degenerate data for fitting")
	}
	a := (float64(n)*sumXY - sumX*sumY) / denom
	b := (sumY - a*sumX) / float64(n)

	// 计算 R² (拟合优度)
	meanY := sumY / float64(n)
	var ssTot, ssRes float64
	for _, p := range points {
		if p.Status == "completed" && len(p.CollectedData) > 0 {
			y := p.CollectedData[0]
			yPred := a*p.TargetPressure + b
			ssTot += (y - meanY) * (y - meanY)
			ssRes += (y - yPred) * (y - yPred)
		}
	}
	r2 := 0.0
	if ssTot > 1e-10 {
		r2 = 1 - ssRes/ssTot
	}

	result := &FittingResult{
		Slope:     a,
		Intercept: b,
		R2:        r2,
		Points:    n,
	}

	s.publish("fitting.completed", map[string]any{
		"slope":     a,
		"intercept": b,
		"r2":        r2,
		"points":    n,
	})

	// 状态迁移: fitting -> completed
	if err := s.sessionMachine.Transition(domain.SessionStateCompleted); err != nil {
		return nil, fmt.Errorf("transition to completed: %w", err)
	}
	s.publishSessionState()

	return result, nil
}

// ManualPressurize 手动打压：设置目标压力并启动压力控制。
func (s *Service) ManualPressurize(ctx context.Context, target float64) error {
	s.mu.Lock()
	s.syncDriversFromSessionLocked()
	if s.pressureDriver == nil {
		s.mu.Unlock()
		return ErrPressureDeviceNotSet
	}
	s.mu.Unlock()

	if err := s.pressureDriver.SetTargetPressure(ctx, target); err != nil {
		return fmt.Errorf("set target pressure: %w", err)
	}

	if ctrl, ok := s.pressureDriver.(interface{ StartControl(context.Context) error }); ok {
		if err := ctrl.StartControl(ctx); err != nil {
			return fmt.Errorf("start pressure control: %w", err)
		}
	}

	return nil
}

// ManualCollect 手动采集：对当前压力点执行一次完整采集（含多样本平均）。
func (s *Service) ManualCollect(ctx context.Context) ([]float64, error) {
	s.mu.Lock()
	s.syncDriversFromSessionLocked()
	if s.measureDriver == nil {
		s.mu.Unlock()
		return nil, ErrMeasureDeviceNotSet
	}
	// 找到下一个 pending 点作为当前采集点
	pointIndex := 0
	for i, p := range s.pressurePoints {
		if p.Status == "pending" {
			pointIndex = i + 1
			break
		}
	}
	if pointIndex == 0 {
		s.mu.Unlock()
		return nil, fmt.Errorf("no pending pressure point to collect")
	}
	s.mu.Unlock()

	return s.Collect(ctx, pointIndex)
}

func averageSamples(samples [][]float64) []float64 {
	if len(samples) == 0 {
		return nil
	}
	if len(samples) == 1 {
		return samples[0]
	}

	// 找到最短的样本长度
	minLen := len(samples[0])
	for _, s := range samples[1:] {
		if len(s) < minLen {
			minLen = len(s)
		}
	}

	result := make([]float64, minLen)
	for i := 0; i < minLen; i++ {
		var sum float64
		for _, s := range samples {
			sum += s[i]
		}
		result[i] = sum / float64(len(samples))
	}

	return result
}
