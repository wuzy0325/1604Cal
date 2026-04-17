package calibration

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"cal1604/internal/device"
	"cal1604/internal/domain"
	"cal1604/internal/infrastructure/driver"
	"cal1604/internal/workflow"
)

// PressurePoint 表示一个压力点及其采集状态。
type PressurePoint struct {
	Index          int       `json:"index"`
	TargetPressure float64   `json:"targetPressure"`
	Status         string    `json:"status"` // pending, pressurizing, stabilizing, collecting, completed, error
	CollectedData  []float64 `json:"collectedData,omitempty"`
	ActualPressure float64   `json:"actualPressure,omitempty"`
}

// CalibrationConfig 校准配置。
type CalibrationConfig struct {
	Channels       []int   `json:"channels"`
	PressurePoints int     `json:"pressurePoints"`
	AverageCount   int     `json:"averageCount"`
	MinPressure    float64 `json:"minPressure"`
	MaxPressure    float64 `json:"maxPressure"`
	StableWaitMs   int     `json:"stableWaitMs"`
}

// CalibrationResult 校准结果。
type CalibrationResult struct {
	Success       bool                `json:"success"`
	State         domain.SessionState `json:"state"`
	CollectedData map[int][]float64   `json:"collectedData,omitempty"` // pointIndex -> channelData
	Error         string              `json:"error,omitempty"`
}

// StatusPublisher 广播事件。
type StatusPublisher func(eventType string, data any)

// Service 校准流程编排服务。
type Service struct {
	mu             sync.Mutex
	sessionMachine *workflow.SessionMachine
	factory        *driver.Factory
	deviceManager  device.DeviceStore
	driverProvider device.ActiveDriverProvider

	measureDriver  device.MeasureDriver
	pressureDriver device.PressureDriver
	measureDevID   string
	pressureDevID  string

	config         CalibrationConfig
	pressurePoints []PressurePoint
	currentPoint   int

	publish StatusPublisher
}

// NewService 创建校准服务。
func NewService(
	sessionMachine *workflow.SessionMachine,
	factory *driver.Factory,
	deviceManager device.DeviceStore,
	publisher StatusPublisher,
	driverProvider device.ActiveDriverProvider,
) *Service {
	if publisher == nil {
		publisher = func(string, any) {}
	}
	return &Service{
		sessionMachine: sessionMachine,
		factory:        factory,
		deviceManager:  deviceManager,
		publish:        publisher,
		driverProvider: driverProvider,
	}
}

// SetMeasureDevice 仅设置计量设备驱动，用于连接后立即读取阀门/单位/设备信息。
// 后续 SetDevices 调用会覆盖此驱动实例。
func (s *Service) SetMeasureDevice(measureDevID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	mDrv, err := s.resolveMeasureDriver(measureDevID)
	if err != nil {
		return err
	}

	s.measureDevID = measureDevID
	s.measureDriver = mDrv
	return nil
}

// SetDevices 设置校准使用的设备。
func (s *Service) SetDevices(measureDevID, pressureDevID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	mDrv, err := s.resolveMeasureDriver(measureDevID)
	if err != nil {
		return err
	}
	pDrv, err := s.resolvePressureDriver(pressureDevID)
	if err != nil {
		return err
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
func (s *Service) GeneratePressurePoints() []PressurePoint {
	s.mu.Lock()
	defer s.mu.Unlock()

	points := s.config.PressurePoints
	if points < 2 {
		points = 2
	}
	if points > 11 {
		points = 11
	}

	minP := s.config.MinPressure
	maxP := s.config.MaxPressure
	if maxP <= minP {
		maxP = minP + 100
	}

	step := (maxP - minP) / float64(points-1)
	s.pressurePoints = make([]PressurePoint, points)
	for i := 0; i < points; i++ {
		s.pressurePoints[i] = PressurePoint{
			Index:          i + 1,
			TargetPressure: minP + step*float64(i),
			Status:         "pending",
		}
	}
	s.currentPoint = 0

	return s.pressurePoints
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
	if s.pressureDriver == nil {
		s.mu.Unlock()
		return fmt.Errorf("pressure device not set")
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

// Collect 从计量设备采集数据。
func (s *Service) Collect(ctx context.Context, pointIndex int) ([]float64, error) {
	s.mu.Lock()
	if s.measureDriver == nil {
		s.mu.Unlock()
		return nil, fmt.Errorf("measure device not set")
	}
	if pointIndex < 1 || pointIndex > len(s.pressurePoints) {
		s.mu.Unlock()
		return nil, fmt.Errorf("invalid point index: %d", pointIndex)
	}
	point := &s.pressurePoints[pointIndex-1]
	point.Status = "collecting"
	channels := s.config.Channels
	avgCount := s.config.AverageCount
	if avgCount < 1 {
		avgCount = 1
	}
	s.mu.Unlock()

	// 状态迁移: -> collecting
	if err := s.sessionMachine.Transition(domain.SessionStateCollecting); err != nil {
		// 可能已经在 collecting
	}
	s.publishSessionState()

	// 多次采集求平均
	allSamples := make([][]float64, 0, avgCount)
	for i := 0; i < avgCount; i++ {
		data, err := s.measureDriver.CollectData(ctx, channels)
		if err != nil {
			s.markPointError(pointIndex, err.Error())
			return nil, fmt.Errorf("collect sample %d: %w", i+1, err)
		}
		allSamples = append(allSamples, data)
		time.Sleep(100 * time.Millisecond)
	}

	// 计算平均值
	averaged := averageSamples(allSamples)

	s.mu.Lock()
	point.CollectedData = averaged
	point.Status = "completed"
	s.currentPoint = pointIndex
	s.mu.Unlock()

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
func (s *Service) Fit(ctx context.Context) error {
	s.mu.Lock()
	if s.measureDriver == nil {
		s.mu.Unlock()
		return fmt.Errorf("measure device not set")
	}
	s.mu.Unlock()

	// 状态迁移: -> fitting
	if err := s.sessionMachine.Transition(domain.SessionStateFitting); err != nil {
		return fmt.Errorf("transition to fitting: %w", err)
	}
	s.publishSessionState()

	// WTN1604 执行拟合
	wtn, ok := s.measureDriver.(*driver.WTN1604Driver)
	if !ok {
		// 如果不是 WTN1604，使用软件拟合
		return s.softwareFit(ctx)
	}

	if err := wtn.PerformFitting(ctx); err != nil {
		return fmt.Errorf("perform fitting: %w", err)
	}

	if err := wtn.SaveCoefficients(ctx); err != nil {
		return fmt.Errorf("save coefficients: %w", err)
	}

	// 状态迁移: fitting -> completed
	if err := s.sessionMachine.Transition(domain.SessionStateCompleted); err != nil {
		return fmt.Errorf("transition to completed: %w", err)
	}
	s.publishSessionState()

	return nil
}

// softwareFit 软件侧拟合（非 WTN1604 设备的备选方案）。
func (s *Service) softwareFit(ctx context.Context) error {
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
		return fmt.Errorf("not enough data points for fitting: %d", n)
	}

	// y = a*x + b
	denom := float64(n)*sumX2 - sumX*sumX
	if math.Abs(denom) < 1e-10 {
		return fmt.Errorf("degenerate data for fitting")
	}
	a := (float64(n)*sumXY - sumX*sumY) / denom
	b := (sumY - a*sumX) / float64(n)

	s.publish("fitting.completed", map[string]any{
		"slope":     a,
		"intercept": b,
		"points":    n,
	})

	// 状态迁移: fitting -> completed
	if err := s.sessionMachine.Transition(domain.SessionStateCompleted); err != nil {
		return fmt.Errorf("transition to completed: %w", err)
	}
	s.publishSessionState()

	return nil
}

// StartCalibration 开始校准流程（WTN1604 多点校准模式）。
func (s *Service) StartCalibration(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.measureDriver == nil || s.pressureDriver == nil {
		return fmt.Errorf("devices not set")
	}

	// 状态迁移: idle -> ready -> pressurizing
	if s.sessionMachine.State() == domain.SessionStateIdle {
		if err := s.sessionMachine.Transition(domain.SessionStateReady); err != nil {
			return fmt.Errorf("transition to ready: %w", err)
		}
		s.publishSessionState()
	}

	// 设置阀门为校准状态
	if err := s.measureDriver.SetValveStatus(ctx, "calibration"); err != nil {
		return fmt.Errorf("set valve to calibration: %w", err)
	}

	// WTN1604 开始多点校准
	wtn, ok := s.measureDriver.(*driver.WTN1604Driver)
	if ok {
		avgCount := s.config.AverageCount
		if avgCount < 1 {
			avgCount = 1
		}
		if err := wtn.StartCalibration(ctx, s.config.Channels, s.config.PressurePoints, avgCount); err != nil {
			return fmt.Errorf("start WTN1604 calibration: %w", err)
		}
	}

	return nil
}

// EndCalibration 结束校准流程。
func (s *Service) EndCalibration(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// WTN1604 结束校准
	wtn, ok := s.measureDriver.(*driver.WTN1604Driver)
	if ok {
		_ = wtn.EndCalibration(ctx)
	}

	// 设置阀门为测量状态
	if s.measureDriver != nil {
		_ = s.measureDriver.SetValveStatus(ctx, "measurement")
	}

	// 停止压力
	if s.pressureDriver != nil {
		_ = s.pressureDriver.Stop(ctx)
	}

	return nil
}

// ReadCurrentPressure 读取打压设备当前压力。
func (s *Service) ReadCurrentPressure(ctx context.Context) (float64, error) {
	s.mu.Lock()
	drv := s.pressureDriver
	s.mu.Unlock()

	if drv == nil {
		return 0, fmt.Errorf("pressure device not set")
	}
	return drv.ReadCurrentPressure(ctx)
}

// ReadStability 读取打压设备稳定状态。
func (s *Service) ReadStability(ctx context.Context) (bool, error) {
	s.mu.Lock()
	drv := s.pressureDriver
	s.mu.Unlock()

	if drv == nil {
		return false, fmt.Errorf("pressure device not set")
	}
	return drv.ReadStability(ctx)
}

// ReadMeasureData 从计量设备读取实时数据。
func (s *Service) ReadMeasureData(ctx context.Context) ([]float64, error) {
	s.mu.Lock()
	drv := s.measureDriver
	channels := s.config.Channels
	s.mu.Unlock()

	if drv == nil {
		return nil, fmt.Errorf("measure device not set")
	}
	if len(channels) == 0 {
		// 默认读取所有16通道
		channels = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	}
	return drv.CollectData(ctx, channels)
}

// ReadValveStatus 读取计量设备阀门状态。
func (s *Service) ReadValveStatus(ctx context.Context) (string, error) {
	s.mu.Lock()
	drv := s.measureDriver
	s.mu.Unlock()

	if drv == nil {
		return "", fmt.Errorf("measure device not set")
	}
	return drv.ReadValveStatus(ctx)
}

// SetValveStatus 设置计量设备阀门状态。
func (s *Service) SetValveStatus(ctx context.Context, status string) error {
	s.mu.Lock()
	drv := s.measureDriver
	s.mu.Unlock()

	if drv == nil {
		return fmt.Errorf("measure device not set")
	}
	return drv.SetValveStatus(ctx, status)
}

// ReadMeasureUnit 读取计量设备压力单位。
func (s *Service) ReadMeasureUnit(ctx context.Context) (string, error) {
	s.mu.Lock()
	drv := s.measureDriver
	s.mu.Unlock()

	if drv == nil {
		return "", fmt.Errorf("measure device not set")
	}
	return drv.ReadUnit(ctx)
}

// SetMeasureUnit 设置计量设备压力单位。
func (s *Service) SetMeasureUnit(ctx context.Context, unit string) error {
	s.mu.Lock()
	drv := s.measureDriver
	s.mu.Unlock()

	if drv == nil {
		return fmt.Errorf("measure device not set")
	}
	return drv.SetUnit(ctx, unit)
}

// ReadDeviceInfo 读取计量设备信息（型号、版本等）。
func (s *Service) ReadDeviceInfo(ctx context.Context) (map[string]string, error) {
	s.mu.Lock()
	drv := s.measureDriver
	s.mu.Unlock()

	if drv == nil {
		return nil, fmt.Errorf("measure device not set")
	}
	return drv.ReadDeviceInfo(ctx)
}

// ResetMeasureDevice 复位计量设备。
func (s *Service) ResetMeasureDevice(ctx context.Context) error {
	s.mu.Lock()
	drv := s.measureDriver
	s.mu.Unlock()

	if drv == nil {
		return fmt.Errorf("measure device not set")
	}
	return drv.Reset(ctx)
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

func (s *Service) markPointError(pointIndex int, reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if pointIndex >= 1 && pointIndex <= len(s.pressurePoints) {
		s.pressurePoints[pointIndex-1].Status = "error"
	}
}

func (s *Service) publishSessionState() {
	s.publish("session.state.changed", map[string]any{
		"state": string(s.sessionMachine.State()),
	})
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
