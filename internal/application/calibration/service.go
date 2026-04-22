package calibration

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"cal1604/internal/application/session"
	"cal1604/internal/device"
	"cal1604/internal/domain"
	"cal1604/internal/infrastructure/driver"
	"cal1604/internal/workflow"
)

// defaultAlarmService 用于报警判定的服务实例。
var defaultAlarmService = workflow.NewAlarmService()

// PressurePoint 表示一个压力点及其采集状态。
type PressurePoint struct {
	Index          int       `json:"index"`
	TargetPressure float64   `json:"targetPressure"`
	Status         string    `json:"status"`    // pending, pressurizing, stabilizing, collecting, completed, error
	Direction      string    `json:"direction"` // forward | backward（回程模式标记正程/回程）
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
	ControlMode    string  `json:"controlMode"`    // auto | manual
	Precision      int     `json:"precision"`      // 小数位数
	PrecisionLevel float64 `json:"precisionLevel"` // 精度等级（如 0.0005）
	PressureMode   string  `json:"pressureMode"`   // single | roundTrip
}

// CalibrationSession 校准会话，记录一次完整校准流程的全部状态。
type CalibrationSession struct {
	ID               string            `json:"id"`
	StartTime        time.Time         `json:"startTime"`
	EndTime          *time.Time        `json:"endTime,omitempty"`
	Config           CalibrationConfig `json:"config"`
	PressurePoints   []PressurePoint   `json:"pressurePoints"`
	MeasureDeviceID  string            `json:"measureDeviceId"`
	PressureDeviceID string            `json:"pressureDeviceId"`
	Status           string            `json:"status"` // running | completed | error
}

// FittingResult 拟合结果。
type FittingResult struct {
	Slope     float64 `json:"slope"`     // 斜率
	Intercept float64 `json:"intercept"` // 截距
	R2        float64 `json:"r2"`        // 拟合优度 R²
	Points    int     `json:"points"`    // 参与拟合的数据点数
}

// CalibrationResult 校准结果。
type CalibrationResult struct {
	Success       bool                `json:"success"`
	State         domain.SessionState `json:"state"`
	CollectedData map[int][]float64   `json:"collectedData,omitempty"` // pointIndex -> channelData
	Error         string              `json:"error,omitempty"`
}

var (
	// ErrMeasureDeviceNotSet 表示计量设备驱动尚未绑定。
	ErrMeasureDeviceNotSet = errors.New("measure device not set")
	// ErrPressureDeviceNotSet 表示打压设备驱动尚未绑定。
	ErrPressureDeviceNotSet = errors.New("pressure device not set")
	// ErrInvalidStartState 表示当前会话状态不允许开始标定。
	ErrInvalidStartState = errors.New("invalid session state for start")
	// ErrNoPendingAlarm 表示当前没有待处理的报警。
	ErrNoPendingAlarm = errors.New("no pending alarm")
	// ErrAutoCollectionRunning 表示自动采集已在运行中。
	ErrAutoCollectionRunning = errors.New("auto collection already running")
)

// StatusPublisher 广播事件。
type StatusPublisher func(eventType string, data any)

// StartPrerequisiteConfig 定义标定启动门禁配置。
type StartPrerequisiteConfig struct {
	EnforceValveCalibration bool
}

func defaultStartPrerequisiteConfig() StartPrerequisiteConfig {
	// 默认关闭阀门门禁，便于设备阀门状态异常时持续联调后续流程。
	return StartPrerequisiteConfig{EnforceValveCalibration: false}
}

// Service 校准流程编排服务。
type Service struct {
	mu             sync.Mutex
	sessionMachine *workflow.SessionMachine
	factory        *driver.Factory
	deviceManager  device.DeviceStore
	driverProvider device.ActiveDriverProvider
	sessionService *session.Service

	measureDriver  device.MeasureDriver
	pressureDriver device.PressureDriver
	measureDevID   string
	pressureDevID  string

	config             CalibrationConfig
	alarmConfig        domain.AlarmConfig
	pressurePoints     []PressurePoint
	currentPoint       int
	calibrationSession *CalibrationSession

	// autoCollectionCtx 用于控制自动采集 goroutine 的生命周期。
	autoCollectionCtx    context.Context
	autoCollectionCancel context.CancelFunc
	autoCollectionMu     sync.Mutex

	// alarmCh 用于在自动采集过程中阻塞等待用户确认报警。
	alarmMu      sync.Mutex
	alarmCh      chan string
	alarmPending bool

	publish StatusPublisher

	startPrerequisiteConfig StartPrerequisiteConfig
}

// NewService 创建校准服务。
func NewService(
	sessionMachine *workflow.SessionMachine,
	factory *driver.Factory,
	deviceManager device.DeviceStore,
	publisher StatusPublisher,
	driverProvider device.ActiveDriverProvider,
	sessionSvc *session.Service,
) *Service {
	if publisher == nil {
		publisher = func(string, any) {}
	}
	return &Service{
		sessionMachine:          sessionMachine,
		factory:                 factory,
		deviceManager:           deviceManager,
		publish:                 publisher,
		driverProvider:          driverProvider,
		sessionService:          sessionSvc,
		startPrerequisiteConfig: defaultStartPrerequisiteConfig(),
	}
}

// SetStartPrerequisiteConfig 设置标定启动门禁配置。
func (s *Service) SetStartPrerequisiteConfig(cfg StartPrerequisiteConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.startPrerequisiteConfig = cfg
}

// syncDriversFromSessionLocked 在持有 mu 锁时，从共享会话同步当前驱动引用。
// 这样标定流程不依赖独立的设备绑定端点，避免出现会话已绑定但标定服务仍为空指针的问题。
func (s *Service) syncDriversFromSessionLocked() {
	if s.sessionService == nil {
		return
	}
	if drv := s.sessionService.MeasureDriver(); drv != nil {
		s.measureDriver = drv
	}
	if drv := s.sessionService.PressureDriver(); drv != nil {
		s.pressureDriver = drv
	}
}

// StartCalibration 开始校准流程（WTN1604 多点校准模式）。
// 若 ControlMode 为 auto，则在准备完成后自动启动采集循环。
// 是否校验阀门状态由 startPrerequisiteConfig.EnforceValveCalibration 控制。
func (s *Service) StartCalibration(ctx context.Context) error {
	s.mu.Lock()
	s.syncDriversFromSessionLocked()
	if s.measureDriver == nil {
		s.mu.Unlock()
		return ErrMeasureDeviceNotSet
	}
	enforceValveGate := s.startPrerequisiteConfig.EnforceValveCalibration

	// 状态迁移: (idle|stopped|completed) -> ready
	state := s.sessionMachine.State()
	switch state {
	case domain.SessionStateIdle, domain.SessionStateStopped, domain.SessionStateCompleted:
		if err := s.sessionMachine.Transition(domain.SessionStateReady); err != nil {
			s.mu.Unlock()
			return fmt.Errorf("transition to ready: %w", err)
		}
		s.publishSessionState()
	case domain.SessionStateReady:
		// 已就绪，保持当前状态
	default:
		s.mu.Unlock()
		return fmt.Errorf("%w: %s", ErrInvalidStartState, state)
	}

	if enforceValveGate {
		valveStatus, err := s.measureDriver.ReadValveStatus(ctx)
		if err != nil {
			s.mu.Unlock()
			return fmt.Errorf("read valve status: %w", err)
		}
		if valveStatus != "calibration" {
			s.mu.Unlock()
			return fmt.Errorf("valve must be in calibration state, current: %s", valveStatus)
		}
	}

	// WTN1604 开始多点校准
	wtn, ok := s.measureDriver.(*driver.WTN1604Driver)
	if ok {
		avgCount := s.config.AverageCount
		if avgCount < 1 {
			avgCount = 1
		}
		if err := wtn.StartCalibration(ctx, s.config.Channels, s.config.PressurePoints, avgCount); err != nil {
			s.mu.Unlock()
			return fmt.Errorf("start WTN1604 calibration: %w", err)
		}
	}

	// 初始化校准会话记录
	s.calibrationSession = &CalibrationSession{
		ID:               fmt.Sprintf("cal-%d", time.Now().UnixMilli()),
		StartTime:        time.Now(),
		Config:           s.config,
		PressurePoints:   s.pressurePoints,
		MeasureDeviceID:  s.measureDevID,
		PressureDeviceID: s.pressureDevID,
		Status:           "running",
	}

	controlMode := s.config.ControlMode
	s.mu.Unlock()

	if controlMode == "auto" {
		if err := s.RunAutoCollection(ctx); err != nil {
			return fmt.Errorf("start auto collection: %w", err)
		}
	}

	return nil
}

// ValidateStartPrerequisites 校验标定启动前置条件是否满足。
// 检查项：计量设备已绑定、通道已选择、配置有效（阀门门禁按配置可开关）。
// 用于 session/start 端点在状态迁移前进行门禁拦截。
func (s *Service) ValidateStartPrerequisites(ctx context.Context) error {
	s.mu.Lock()
	s.syncDriversFromSessionLocked()
	measureDriver := s.measureDriver
	pressureDriver := s.pressureDriver
	channels := s.config.Channels
	config := s.config
	enforceValveGate := s.startPrerequisiteConfig.EnforceValveCalibration
	s.mu.Unlock()

	if measureDriver == nil {
		return fmt.Errorf("measure device not bound")
	}

	if len(channels) == 0 {
		return fmt.Errorf("no channels selected")
	}

	if config.PressurePoints < 2 || config.PressurePoints > 6 {
		return fmt.Errorf("pressure points must be between 2 and 6, got %d", config.PressurePoints)
	}

	// 按模式化门禁校验打压设备
	if config.ControlMode == "auto" && pressureDriver == nil {
		return fmt.Errorf("auto mode requires pressure device to be bound")
	}

	if enforceValveGate {
		valveStatus, err := measureDriver.ReadValveStatus(ctx)
		if err != nil {
			return fmt.Errorf("read valve status: %w", err)
		}
		if valveStatus != "calibration" {
			return fmt.Errorf("valve must be in calibration state, current: %s", valveStatus)
		}
	}

	return nil
}

// EndCalibration 结束校准流程，执行确定性资源清理。
// 停止自动采集循环、停止压力控制、结束 WTN1604 校准。
// 不自动切阀（保留人工回阀路径），状态迁移由 handler 层完成。
func (s *Service) EndCalibration(ctx context.Context) error {
	s.StopAutoCollection()

	s.mu.Lock()
	defer s.mu.Unlock()
	s.syncDriversFromSessionLocked()

	// WTN1604 结束校准
	wtn, ok := s.measureDriver.(*driver.WTN1604Driver)
	if ok {
		_ = wtn.EndCalibration(ctx)
	}

	// 停止压力控制
	if s.pressureDriver != nil {
		_ = s.pressureDriver.Stop(ctx)
	}

	// 关闭校准会话记录
	if s.calibrationSession != nil {
		now := time.Now()
		s.calibrationSession.EndTime = &now
		s.calibrationSession.Status = "completed"
		s.calibrationSession.PressurePoints = s.pressurePoints
	}

	return nil
}

// RunAutoCollection 启动自动采集流程，按压力点列表依次执行打压、稳定、采集。
// 该方法在后台 goroutine 中运行，通过 context 控制取消。
func (s *Service) RunAutoCollection(ctx context.Context) error {
	s.autoCollectionMu.Lock()
	if s.autoCollectionCancel != nil {
		s.autoCollectionMu.Unlock()
		return ErrAutoCollectionRunning
	}
	s.autoCollectionCtx, s.autoCollectionCancel = context.WithCancel(ctx)
	s.autoCollectionMu.Unlock()

	go func() {
		defer func() {
			s.autoCollectionMu.Lock()
			s.autoCollectionCancel = nil
			s.autoCollectionCtx = nil
			s.autoCollectionMu.Unlock()
		}()

		s.publish("autoCollection.started", map[string]any{
			"pointCount": len(s.pressurePoints),
			"mode":       s.config.ControlMode,
		})

		for i := s.currentPoint; i < len(s.pressurePoints); i++ {
			s.autoCollectionMu.Lock()
			ctx := s.autoCollectionCtx
			s.autoCollectionMu.Unlock()
			if ctx == nil || ctx.Err() != nil {
				break
			}

			s.mu.Lock()
			s.currentPoint = i
			s.mu.Unlock()

			if err := s.collectPoint(ctx, i+1); err != nil {
				s.publish("autoCollection.error", map[string]any{
					"pointIndex": i + 1,
					"error":      err.Error(),
				})
				_ = s.sessionMachine.Transition(domain.SessionStateError)
				s.publishSessionState()
				// 标记会话错误
				if s.calibrationSession != nil {
					s.calibrationSession.Status = "error"
				}
				return
			}
		}

		s.autoCollectionMu.Lock()
		ctx := s.autoCollectionCtx
		s.autoCollectionMu.Unlock()
		if ctx != nil && ctx.Err() == nil {
			s.publish("autoCollection.completed", map[string]any{
				"totalPoints": len(s.pressurePoints),
			})
		}
	}()

	return nil
}

// collectPoint 采集单个压力点：打压 -> 稳定监控 -> 采集 -> 报警检查。
func (s *Service) collectPoint(ctx context.Context, pointIndex int) error {
	s.publish("point.started", map[string]any{"pointIndex": pointIndex})

	if err := s.Pressurize(ctx, pointIndex); err != nil {
		return fmt.Errorf("pressurize point %d: %w", pointIndex, err)
	}

	// 稳定性监控：使用 StabilityMonitor 进行带 SSE 事件推送的稳定判定
	s.mu.Lock()
	point := s.pressurePoints[pointIndex-1]
	tolerance := s.config.PrecisionLevel
	stableDurationMs := s.config.StableWaitMs
	s.mu.Unlock()

	if tolerance <= 0 {
		tolerance = 0.0005
	}
	if stableDurationMs <= 0 {
		stableDurationMs = 5000
	}

	monitor := workflow.NewStabilityMonitor(
		tolerance,
		time.Duration(stableDurationMs)*time.Millisecond,
		workflow.StabilityEventPublisher(s.publish),
	)

	// 循环读取压力直到稳定
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		actual, err := s.pressureDriver.ReadCurrentPressure(ctx)
		if err != nil {
			return fmt.Errorf("read pressure for stability check: %w", err)
		}

		status := monitor.FeedSample(point.TargetPressure, actual)
		if status.IsStable {
			break
		}

		time.Sleep(200 * time.Millisecond)
	}

	data, err := s.Collect(ctx, pointIndex)
	if err != nil {
		return fmt.Errorf("collect point %d: %w", pointIndex, err)
	}

	// 报警检查
	action, err := s.checkAlarm(ctx, pointIndex, data)
	if err != nil {
		return fmt.Errorf("alarm check point %d: %w", pointIndex, err)
	}
	if action == "retry" {
		s.mu.Lock()
		if pointIndex >= 1 && pointIndex <= len(s.pressurePoints) {
			s.pressurePoints[pointIndex-1].Status = "pending"
			s.pressurePoints[pointIndex-1].CollectedData = nil
		}
		s.mu.Unlock()
		s.publish("point.retry", map[string]any{"pointIndex": pointIndex})
		return s.collectPoint(ctx, pointIndex)
	}

	s.publish("point.completed", map[string]any{"pointIndex": pointIndex, "data": data})
	return nil
}

// checkAlarm 检查采集数据是否触发报警，若触发则阻塞等待用户决策。
// 返回决策动作："continue" 或 "retry"。
func (s *Service) checkAlarm(ctx context.Context, pointIndex int, data []float64) (string, error) {
	s.mu.Lock()
	point := s.pressurePoints[pointIndex-1]
	alarmConfig := s.alarmConfig
	channels := s.config.Channels
	maxPressure := s.config.MaxPressure
	s.mu.Unlock()

	if len(data) == 0 {
		return "continue", nil
	}

	// 多通道报警判定
	channelData := make(map[int]float64)
	for i, ch := range channels {
		if i < len(data) {
			channelData[ch] = data[i]
		}
	}

	result := defaultAlarmService.EvaluateMultiChannel(alarmConfig, point.TargetPressure, maxPressure, channelData)
	if !result.Triggered {
		return "continue", nil
	}

	// 触发报警，进入等待报警确认状态
	if err := s.sessionMachine.Transition(domain.SessionStateAwaitAlarmResolution); err != nil {
		// 若状态机不支持该迁移，直接继续
		return "continue", nil
	}
	s.publishSessionState()

	s.alarmMu.Lock()
	s.alarmCh = make(chan string, 1)
	s.alarmPending = true
	alarmCh := s.alarmCh
	s.alarmMu.Unlock()

	s.publish("alarm.triggered", map[string]any{
		"pointIndex":        pointIndex,
		"targetPressure":    point.TargetPressure,
		"overLimitChannels": result.OverLimitChannels,
		"maxDeviation":      result.MaxDeviation,
		"channelDetails":    result.ChannelDetails,
	})

	select {
	case decision := <-alarmCh:
		s.alarmMu.Lock()
		s.alarmPending = false
		s.alarmCh = nil
		s.alarmMu.Unlock()

		// 从报警状态恢复
		_ = s.sessionMachine.Transition(domain.SessionStateCollecting)
		s.publishSessionState()

		if decision == "retry" {
			return "retry", nil
		}
		return "continue", nil
	case <-ctx.Done():
		s.alarmMu.Lock()
		s.alarmPending = false
		s.alarmCh = nil
		s.alarmMu.Unlock()
		return "", ctx.Err()
	}
}

// ResolveAlarm 用户确认报警，传入决策动作："continue" 或 "retry"。
func (s *Service) ResolveAlarm(decision string) error {
	s.alarmMu.Lock()
	defer s.alarmMu.Unlock()

	if !s.alarmPending || s.alarmCh == nil {
		return ErrNoPendingAlarm
	}

	if decision != "continue" && decision != "retry" {
		return fmt.Errorf("invalid alarm decision: %s", decision)
	}

	select {
	case s.alarmCh <- decision:
		return nil
	default:
		return fmt.Errorf("alarm channel blocked")
	}
}

// RetryPoint 重试指定压力点，将其状态重置为 pending 并重新采集。
func (s *Service) RetryPoint(ctx context.Context, pointIndex int) error {
	s.mu.Lock()
	if pointIndex < 1 || pointIndex > len(s.pressurePoints) {
		s.mu.Unlock()
		return fmt.Errorf("invalid point index: %d", pointIndex)
	}
	s.pressurePoints[pointIndex-1].Status = "pending"
	s.pressurePoints[pointIndex-1].CollectedData = nil
	s.pressurePoints[pointIndex-1].ActualPressure = 0
	s.currentPoint = pointIndex - 1
	controlMode := s.config.ControlMode
	hasPressureDriver := s.pressureDriver != nil
	s.mu.Unlock()

	s.publish("point.retry", map[string]any{"pointIndex": pointIndex})

	// 手动模式且未连接打压设备时，仅重置为待确认，
	// 由操作者再次确认后执行采集，不自动触发打压链路。
	if controlMode == "manual" && !hasPressureDriver {
		return nil
	}

	return s.collectPoint(ctx, pointIndex)
}

// PauseAutoCollection 暂停自动采集，将状态机迁移到 paused。
func (s *Service) PauseAutoCollection() error {
	if err := s.sessionMachine.Transition(domain.SessionStatePaused); err != nil {
		return fmt.Errorf("transition to paused: %w", err)
	}
	s.publishSessionState()
	return nil
}

// ResumeAutoCollection 恢复自动采集，重新启动从当前点的自动采集循环。
func (s *Service) ResumeAutoCollection(ctx context.Context) error {
	if err := s.sessionMachine.Transition(domain.SessionStatePressurizing); err != nil {
		return fmt.Errorf("transition to pressurizing: %w", err)
	}
	s.publishSessionState()
	return s.RunAutoCollection(ctx)
}

// StopAutoCollection 停止自动采集，取消后台 goroutine。
func (s *Service) StopAutoCollection() {
	s.autoCollectionMu.Lock()
	if s.autoCollectionCancel != nil {
		s.autoCollectionCancel()
		s.autoCollectionCancel = nil
		s.autoCollectionCtx = nil
	}
	s.autoCollectionMu.Unlock()

	s.alarmMu.Lock()
	if s.alarmPending && s.alarmCh != nil {
		close(s.alarmCh)
		s.alarmPending = false
		s.alarmCh = nil
	}
	s.alarmMu.Unlock()
}

// IsAutoCollectionRunning 返回自动采集是否正在运行。
func (s *Service) IsAutoCollectionRunning() bool {
	s.autoCollectionMu.Lock()
	defer s.autoCollectionMu.Unlock()
	return s.autoCollectionCancel != nil
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
