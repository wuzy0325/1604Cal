package calibration

import (
	"context"
	"fmt"
	"time"

	"cal1604/internal/application/session"
	"cal1604/internal/device"
	"cal1604/internal/domain"
	"cal1604/internal/events"
	"cal1604/internal/workflow"
)

// SetDevices 设置校准使用的设备。
// 委托给 session.Service 处理，同时保持本地驱动引用以供标定流程使用。
func (s *Service) SetDevices(measureDevID, pressureDevID string) error {
	if s.sessionService != nil {
		if err := s.sessionService.BindDevices(measureDevID, pressureDevID); err != nil {
			return err
		}
		s.mu.Lock()
		s.measureDevID = measureDevID
		s.pressureDevID = pressureDevID
		s.mu.Unlock()
		return nil
	}

	// 无 sessionService 时，使用 DriverResolver 直接解析驱动。
	s.mu.Lock()
	defer s.mu.Unlock()
	resolver := &session.DriverResolver{
		DeviceManager:  s.deviceManager,
		DriverProvider: s.driverProvider,
		Factory:        s.factory,
	}
	mDrv, err := resolver.ResolveMeasureDriver(measureDevID)
	if err != nil {
		return err
	}
	var pDrv device.PressureDriver
	if pressureDevID != "" {
		pDrv, err = resolver.ResolvePressureDriver(pressureDevID)
		if err != nil {
			return err
		}
	}
	s.measureDevID = measureDevID
	s.pressureDevID = pressureDevID
	s.measureDriver = mDrv
	s.pressureDriver = pDrv
	return nil
}

// SetConfig 设置校准配置。
func (s *Service) SetConfig(config domain.WorkflowConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = config
}

// SetAlarmConfig 设置报警配置。
func (s *Service) SetAlarmConfig(config domain.AlarmConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.alarmConfig = config
}

// GetAlarmConfig 获取当前报警配置。
func (s *Service) GetAlarmConfig() domain.AlarmConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.alarmConfig
}

// GetCalibrationSession 获取当前校准会话（可能为 nil）。
func (s *Service) GetCalibrationSession() *CalibrationSession {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calibrationSession
}

// SetChannels 设置采集通道。
func (s *Service) SetChannels(channels []int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config.Channels = channels
}

// GetChannels 获取当前通道配置。
func (s *Service) GetChannels() []int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.config.Channels
}

// GeneratePressurePoints 根据配置生成压力点。
// 测点数范围统一为 2~6，超出范围返回错误，禁止隐式裁剪。
// 当 PressureMode 为 roundTrip 时，在正程递增点后追加回程递降点（不含重复的极值点）。
func (s *Service) GeneratePressurePoints() ([]domain.PressurePoint, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	points := s.config.PointCount
	if points < 2 || points > 6 {
		return nil, fmt.Errorf("pressure points must be between 2 and 6, got %d", points)
	}

	minP := s.config.MinPressure
	maxP := s.config.MaxPressure
	if maxP <= minP {
		return nil, fmt.Errorf("maxPressure(%v) must be greater than minPressure(%v)", maxP, minP)
	}

	prec := s.config.Precision
	if prec < 0 {
		prec = 0
	}

	roundTrip := s.config.PressureMode == "roundTrip"
	s.pressurePoints = domain.EquidistantPoints(minP, maxP, points, prec, roundTrip)
	s.currentPoint = 0

	return s.pressurePoints, nil
}

// GetPressurePoints 获取当前压力点列表。
func (s *Service) GetPressurePoints() []domain.PressurePoint {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]domain.PressurePoint, len(s.pressurePoints))
	copy(result, s.pressurePoints)
	return result
}

// Pressurize 对指定压力点执行打压。
func (s *Service) Pressurize(ctx context.Context, pointIndex int) error {
	s.mu.Lock()
	pressureDriver := s.getPressureDriver()
	if pressureDriver == nil {
		s.mu.Unlock()
		return ErrPressureDeviceNotSet
	}
	if pointIndex < 1 || pointIndex > len(s.pressurePoints) {
		s.mu.Unlock()
		return fmt.Errorf("invalid point index: %d", pointIndex)
	}
	targetPressure := s.pressurePoints[pointIndex-1].TargetPressure
	s.mu.Unlock()

	s.updatePointStatus(pointIndex, domain.PointStatusPressurizing)

	// 状态迁移: -> pressurizing
	if err := s.sessionMachine.Transition(domain.SessionStatePressurizing); err != nil {
		return fmt.Errorf("transition to pressurizing: %w", err)
	}
	s.publishSessionState()

	// 设置目标压力
	if err := pressureDriver.SetTargetPressure(ctx, targetPressure); err != nil {
		s.markPointError(pointIndex, err.Error())
		return fmt.Errorf("set target pressure: %w", err)
	}

	// 启动压力控制
	if ctrl, ok := pressureDriver.(device.PressureControlCapable); ok {
		if err := ctrl.StartControl(ctx); err != nil {
			s.markPointError(pointIndex, err.Error())
			return fmt.Errorf("start pressure control: %w", err)
		}
	}

	s.updatePointStatus(pointIndex, domain.PointStatusStabilizing)

	// 状态迁移: pressurizing -> stabilizing
	if err := s.sessionMachine.Transition(domain.SessionStateStabilizing); err != nil {
		// 可能已经在 stabilizing，忽略
	}
	s.publishSessionState()

	// 等待压力稳定
	if err := s.waitForStabilityWithMonitor(ctx, targetPressure); err != nil {
		return fmt.Errorf("wait for stability: %w", err)
	}

	// 读取实际压力
	actual, err := s.getPressureDriver().ReadCurrentPressure(ctx)
	if err == nil {
		s.mu.Lock()
		s.pressurePoints[pointIndex-1].ActualPressure = actual
		s.mu.Unlock()
	}

	s.publish(events.EventPressureApplied, map[string]any{
		"pointIndex":     pointIndex,
		"targetPressure": targetPressure,
		"actualPressure": actual,
	})

	return nil
}

// waitForStabilityWithMonitor 使用 StabilityMonitor 等待压力稳定。
func (s *Service) waitForStabilityWithMonitor(ctx context.Context, targetPressure float64) error {
	tolerance := s.config.PrecisionLevel
	if tolerance <= 0 {
		tolerance = 0.0005
	}
	stableWaitMs := s.config.StableWaitMs
	if stableWaitMs <= 0 {
		stableWaitMs = 5000
	}
	monitor := workflow.NewStabilityMonitor(
		tolerance,
		time.Duration(stableWaitMs)*time.Millisecond,
		workflow.StabilityEventPublisher(s.publish),
	)

	deadline := time.Now().Add(60 * time.Second)
	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("stability timeout")
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		actual, err := s.getPressureDriver().ReadCurrentPressure(ctx)
		if err != nil {
			time.Sleep(time.Duration(stableWaitMs) * time.Millisecond)
			return nil
		}
		status := monitor.FeedSample(targetPressure, actual)
		if status.IsStable {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
}
