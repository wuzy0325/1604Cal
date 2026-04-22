package calibration

import (
	"context"
	"fmt"
	"time"

	"cal1604/internal/device"
	"cal1604/internal/domain"
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
		s.measureDriver = s.sessionService.MeasureDriver()
		s.pressureDriver = s.sessionService.PressureDriver()
		s.mu.Unlock()
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	mDrv, err := s.resolveMeasureDriver(measureDevID)
	if err != nil {
		return err
	}
	var pDrv device.PressureDriver
	if pressureDevID != "" {
		pDrv, err = s.resolvePressureDriver(pressureDevID)
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

// resolveMeasureDriver 优先复用已连接的计量驱动，避免创建未连接的驱动实例。
func (s *Service) resolveMeasureDriver(measureDevID string) (device.MeasureDriver, error) {
	if s.driverProvider != nil {
		if drv := s.driverProvider.GetActiveDriver(measureDevID); drv != nil {
			if mDrv, ok := drv.(device.MeasureDriver); ok {
				return mDrv, nil
			}
		}
	}

	measureDev, ok := s.deviceManager.Get(measureDevID)
	if !ok {
		return nil, fmt.Errorf("measure device %s not found", measureDevID)
	}
	return s.factory.CreateMeasureDriver(measureDev)
}

// resolvePressureDriver 优先复用已连接的打压驱动，避免创建未连接的驱动实例。
func (s *Service) resolvePressureDriver(pressureDevID string) (device.PressureDriver, error) {
	if s.driverProvider != nil {
		if drv := s.driverProvider.GetActiveDriver(pressureDevID); drv != nil {
			if pDrv, ok := drv.(device.PressureDriver); ok {
				return pDrv, nil
			}
		}
	}

	pressureDev, ok := s.deviceManager.Get(pressureDevID)
	if !ok {
		return nil, fmt.Errorf("pressure device %s not found", pressureDevID)
	}
	return s.factory.CreatePressureDriver(pressureDev)
}

// SetConfig 设置校准配置。
func (s *Service) SetConfig(config CalibrationConfig) {
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
func (s *Service) GeneratePressurePoints() ([]PressurePoint, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	points := s.config.PressurePoints
	if points < 2 || points > 6 {
		return nil, fmt.Errorf("pressure points must be between 2 and 6, got %d", points)
	}

	minP := s.config.MinPressure
	maxP := s.config.MaxPressure
	if maxP <= minP {
		maxP = minP + 100
	}

	step := (maxP - minP) / float64(points-1)

	// 正程：递增
	forward := make([]PressurePoint, points)
	for i := 0; i < points; i++ {
		forward[i] = PressurePoint{
			Index:          i + 1,
			TargetPressure: minP + step*float64(i),
			Status:         "pending",
			Direction:      "forward",
		}
	}

	if s.config.PressureMode != "roundTrip" {
		s.pressurePoints = forward
		s.currentPoint = 0
		return s.pressurePoints, nil
	}

	// 回程：递降（不含最后一个正程点，即最大值，避免重复）
	backward := make([]PressurePoint, points-1)
	for i := 0; i < points-1; i++ {
		backward[i] = PressurePoint{
			Index:          points + i + 1,
			TargetPressure: maxP - step*float64(i+1),
			Status:         "pending",
			Direction:      "backward",
		}
	}

	s.pressurePoints = append(forward, backward...)
	s.currentPoint = 0

	return s.pressurePoints, nil
}

// GetPressurePoints 获取当前压力点列表。
func (s *Service) GetPressurePoints() []PressurePoint {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]PressurePoint, len(s.pressurePoints))
	copy(result, s.pressurePoints)
	return result
}

// Pressurize 对指定压力点执行打压。
func (s *Service) Pressurize(ctx context.Context, pointIndex int) error {
	s.mu.Lock()
	s.syncDriversFromSessionLocked()
	if s.pressureDriver == nil {
		s.mu.Unlock()
		return ErrPressureDeviceNotSet
	}
	if pointIndex < 1 || pointIndex > len(s.pressurePoints) {
		s.mu.Unlock()
		return fmt.Errorf("invalid point index: %d", pointIndex)
	}
	point := &s.pressurePoints[pointIndex-1]
	point.Status = "pressurizing"
	s.mu.Unlock()

	// 状态迁移: -> pressurizing
	if err := s.sessionMachine.Transition(domain.SessionStatePressurizing); err != nil {
		return fmt.Errorf("transition to pressurizing: %w", err)
	}
	s.publishSessionState()

	// 设置目标压力
	if err := s.pressureDriver.SetTargetPressure(ctx, point.TargetPressure); err != nil {
		s.markPointError(pointIndex, err.Error())
		return fmt.Errorf("set target pressure: %w", err)
	}

	// 启动压力控制
	if ctrl, ok := s.pressureDriver.(interface{ StartControl(context.Context) error }); ok {
		if err := ctrl.StartControl(ctx); err != nil {
			s.markPointError(pointIndex, err.Error())
			return fmt.Errorf("start pressure control: %w", err)
		}
	}

	s.mu.Lock()
	point.Status = "stabilizing"
	s.mu.Unlock()

	// 状态迁移: pressurizing -> stabilizing
	if err := s.sessionMachine.Transition(domain.SessionStateStabilizing); err != nil {
		// 可能已经在 stabilizing，忽略
	}
	s.publishSessionState()

	// 等待压力稳定
	stableWait := s.config.StableWaitMs
	if stableWait <= 0 {
		stableWait = 2000
	}
	if err := s.waitForStability(ctx, time.Duration(stableWait)*time.Millisecond); err != nil {
		return fmt.Errorf("wait for stability: %w", err)
	}

	// 读取实际压力
	actual, err := s.pressureDriver.ReadCurrentPressure(ctx)
	if err == nil {
		s.mu.Lock()
		point.ActualPressure = actual
		s.mu.Unlock()
	}

	s.publish("pressure.applied", map[string]any{
		"pointIndex":     pointIndex,
		"targetPressure": point.TargetPressure,
		"actualPressure": actual,
	})

	return nil
}

// waitForStability 等待压力稳定。
func (s *Service) waitForStability(ctx context.Context, minStableTime time.Duration) error {
	deadline := time.Now().Add(60 * time.Second)
	stableStart := time.Time{}

	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("stability timeout")
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		stable, err := s.pressureDriver.ReadStability(ctx)
		if err != nil {
			// 如果设备不支持稳定状态查询，使用简单等待
			time.Sleep(minStableTime)
			return nil
		}

		if stable {
			if stableStart.IsZero() {
				stableStart = time.Now()
			}
			if time.Since(stableStart) >= minStableTime {
				return nil
			}
		} else {
			stableStart = time.Time{}
		}

		time.Sleep(200 * time.Millisecond)
	}
}
