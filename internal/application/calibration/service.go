package calibration

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"cal1604/internal/application/session"
	"cal1604/internal/device"
	"cal1604/internal/domain"
	"cal1604/internal/events"
	"cal1604/internal/infrastructure/driver"
	"cal1604/internal/workflow"
)

// defaultAlarmService 用于报警判定的服务实例。
var defaultAlarmService = workflow.NewAlarmService()

// CalibrationSession 是 domain.WorkflowSession 的类型别名，保持外部包向后兼容。
type CalibrationSession = domain.WorkflowSession

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
	// ErrPointSkipped 表示用户选择跳过当前点。
	ErrPointSkipped = errors.New("point skipped by user")
	// ErrInvalidStartState 表示当前会话状态不允许开始标定。
	ErrInvalidStartState = errors.New("invalid session state for start")
	// ErrNoPendingAlarm 表示当前没有待处理的报警。
	ErrNoPendingAlarm = errors.New("no pending alarm")
	// ErrAutoCollectionRunning 表示自动采集已在运行中。
	ErrAutoCollectionRunning = errors.New("auto collection already running")
	// errAutoCollectionStopped 表示自动采集因报警决策主动停止。
	errAutoCollectionStopped = errors.New("auto collection stopped by alarm decision")
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

	config             domain.WorkflowConfig
	alarmConfig        domain.AlarmConfig
	pressurePoints     []domain.PressurePoint
	currentPoint       int
	calibrationSession *CalibrationSession

	// autoCollectionCtx 用于控制自动采集 goroutine 的生命周期。
	autoCollectionCtx    context.Context
	autoCollectionCancel context.CancelFunc
	autoCollectionMu     sync.Mutex
	autoCollectWg        sync.WaitGroup

	// alarmCh 用于在自动采集过程中阻塞等待用户确认报警。
	alarmMu      sync.Mutex
	alarmCh      chan string
	alarmPending bool

	// stabilityTimeoutCh 用于等待前端用户对稳定超时的决定。
	stabilityTimeoutCh chan string

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
		stabilityTimeoutCh:      make(chan string, 1),
		startPrerequisiteConfig: defaultStartPrerequisiteConfig(),
	}
}

// SetStartPrerequisiteConfig 设置标定启动门禁配置。
func (s *Service) SetStartPrerequisiteConfig(cfg StartPrerequisiteConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.startPrerequisiteConfig = cfg
}

// getMeasureDriver 从 session 服务获取计量驱动；session 为 nil 时回退到本地字段。
func (s *Service) getMeasureDriver() device.MeasureDriver {
	if s.sessionService != nil {
		return s.sessionService.MeasureDriver()
	}
	return s.measureDriver
}

// getPressureDriver 从 session 服务获取打压驱动；session 为 nil 时回退到本地字段。
func (s *Service) getPressureDriver() device.PressureDriver {
	if s.sessionService != nil {
		return s.sessionService.PressureDriver()
	}
	return s.pressureDriver
}

// StartCalibration 开始校准流程（WTN1604 多点校准模式）。
// 若 ControlMode 为 auto，则在准备完成后自动启动采集循环。
// 是否校验阀门状态由 startPrerequisiteConfig.EnforceValveCalibration 控制。
func (s *Service) StartCalibration(ctx context.Context) error {
	s.mu.Lock()

	if s.getMeasureDriver() == nil {
		s.mu.Unlock()
		return ErrMeasureDeviceNotSet
	}

	// 在锁内提取所需配置
	enforceValveGate := s.startPrerequisiteConfig.EnforceValveCalibration
	measureDriver := s.getMeasureDriver()
	avgCount := s.config.AverageCount
	channels := s.config.Channels
	pointCount := s.config.PointCount
	if avgCount < 1 {
		avgCount = 1
	}

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

	// 阀门门禁（持有锁时执行，ReadValveStatus 可能 I/O 阻塞，但门禁仅用于开发联调阶段可接受）
	if enforceValveGate {
		valveStatus, err := measureDriver.ReadValveStatus(ctx)
		if err != nil {
			s.mu.Unlock()
			return fmt.Errorf("read valve status: %w", err)
		}
		if valveStatus != "calibration" {
			s.mu.Unlock()
			return fmt.Errorf("valve must be in calibration state, current: %s", valveStatus)
		}
	}

	calDev, calDevOK := measureDriver.(device.CalibrationCapable)
	s.mu.Unlock()

	// WTN1604 开始多点校准（不持有锁，避免 I/O 阻塞影响其他操作）
	if calDevOK {
		if err := calDev.StartCalibration(ctx, channels, pointCount, avgCount); err != nil {
			return fmt.Errorf("start WTN1604 calibration: %w", err)
		}
	}

	// 初始化校准会话记录
	s.mu.Lock()
	s.calibrationSession = &CalibrationSession{
		ID:               fmt.Sprintf("cal-%d", time.Now().UnixMilli()),
		StartTime:        time.Now(),
		Config:           s.config,
		Points:           s.pressurePoints,
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
	
	measureDriver := s.getMeasureDriver()
	pressureDriver := s.getPressureDriver()
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

	if config.PointCount < 2 || config.PointCount > 6 {
		return fmt.Errorf("pressure points must be between 2 and 6, got %d", config.PointCount)
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
// 不自动切阀（保留人工回阀路径），不持有锁时执行设备 I/O。
func (s *Service) EndCalibration(ctx context.Context) error {
	s.StopAutoCollection()

	// 先获取驱动引用（持有锁），I/O 操作在锁外执行
	s.mu.Lock()
	calDev, _ := s.getMeasureDriver().(device.CalibrationCapable)
	pressureDriver := s.getPressureDriver()
	s.mu.Unlock()

	// WTN1604 结束校准（不持有锁，避免 I/O 阻塞影响其他操作）
	if calDev != nil {
		_ = calDev.EndCalibration(ctx)
	}

	// 停止压力控制
	if pressureDriver != nil {
		_ = pressureDriver.Stop(ctx)
	}

	// 关闭校准会话记录
	s.mu.Lock()
	if s.calibrationSession != nil {
		now := time.Now()
		s.calibrationSession.EndTime = &now
		s.calibrationSession.Status = "completed"
		s.calibrationSession.Points = s.pressurePoints
	}
	s.mu.Unlock()

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
	// 使用 context.Background() 而非传入的 ctx（HTTP 请求上下文），
	// 因为 HTTP handler 返回后 r.Context() 会被立即取消，导致自动采集 goroutine 立即退出。
	s.autoCollectionCtx, s.autoCollectionCancel = context.WithCancel(context.Background())
	s.autoCollectionMu.Unlock()

	s.autoCollectWg.Add(1)

	go func() {
		defer s.autoCollectWg.Done()
		defer func() {
			s.autoCollectionMu.Lock()
			s.autoCollectionCancel = nil
			s.autoCollectionCtx = nil
			s.autoCollectionMu.Unlock()
		}()

		s.mu.Lock()
		startIndex := s.resumePointIndexLocked()
		totalPoints := len(s.pressurePoints)
		mode := s.config.ControlMode
		s.currentPoint = startIndex
		s.mu.Unlock()

		s.publish(events.EventAutoCollectionStarted, map[string]any{
			"pointCount": totalPoints,
			"mode":       mode,
			"startIndex": startIndex + 1,
		})

		for i := startIndex; i < totalPoints; i++ {
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
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return
				}

				if errors.Is(err, errAutoCollectionStopped) {
					s.publish(events.EventAutoCollectionStopped, map[string]any{
						"pointIndex": i + 1,
					})
					return
				}

				s.publish(events.EventAutoCollectionError, map[string]any{
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

			s.mu.Lock()
			s.currentPoint = i + 1
			s.mu.Unlock()
		}

		s.autoCollectionMu.Lock()
		ctx := s.autoCollectionCtx
		s.autoCollectionMu.Unlock()
		if ctx != nil && ctx.Err() == nil {
			s.publish(events.EventAutoCollectionCompleted, map[string]any{
				"totalPoints": totalPoints,
			})
		}
	}()

	return nil
}

// collectPoint 采集单个压力点：打压 -> 稳定监控 -> 采集 -> 报警检查。
// 使用迭代循环替代递归，设置最大重试次数避免栈溢出风险。
func (s *Service) collectPoint(ctx context.Context, pointIndex int) error {
	s.publish(events.EventPointStarted, map[string]any{"pointIndex": pointIndex})

	const maxRetries = 3
	for retry := 0; retry <= maxRetries; retry++ {
		if retry > 0 {
			// 重试前重置压力点状态，重新打压
			s.mu.Lock()
			if pointIndex >= 1 && pointIndex <= len(s.pressurePoints) {
				s.pressurePoints[pointIndex-1].Status = domain.PointStatusPending
				s.pressurePoints[pointIndex-1].CollectedData = nil
				s.pressurePoints[pointIndex-1].ActualPressure = 0
			}
			s.mu.Unlock()
			s.publish(events.EventPointRecollect, map[string]any{"pointIndex": pointIndex})
		}

		if err := s.Pressurize(ctx, pointIndex); err != nil {
			if errors.Is(err, ErrPointSkipped) {
				s.updatePointStatus(pointIndex, domain.PointStatusSkipped)
				s.publish(events.EventPointSkipped, map[string]any{"pointIndex": pointIndex})
				return nil
			}
			return fmt.Errorf("pressurize point %d: %w", pointIndex, err)
		}

		// Pressurize() 返回表示打压已完成且压力稳定，直接进入采集。
		data, err := s.Collect(ctx, pointIndex)
		if err != nil {
			return fmt.Errorf("collect point %d: %w", pointIndex, err)
		}

		// 报警检查
		action, err := s.checkAlarm(ctx, pointIndex, data)
		if err != nil {
			return fmt.Errorf("alarm check point %d: %w", pointIndex, err)
		}

		switch action {
		case workflow.AlarmDecisionRecollect:
			// 继续重试循环
			continue

		case workflow.AlarmDecisionSkip:
			s.mu.Lock()
			if pointIndex >= 1 && pointIndex <= len(s.pressurePoints) {
				s.pressurePoints[pointIndex-1].CollectedData = nil
			}
			s.mu.Unlock()
			s.updatePointStatus(pointIndex, domain.PointStatusSkipped)
			s.publish(events.EventPointSkipped, map[string]any{"pointIndex": pointIndex})
			return nil

		case workflow.AlarmDecisionStop:
			s.publish(events.EventPointStopped, map[string]any{"pointIndex": pointIndex})
			return errAutoCollectionStopped
		}

		// AlarmDecisionContinue — 正常完成
		s.publish(events.EventPointCompleted, map[string]any{"pointIndex": pointIndex, "data": data})
		return nil
	}

	return fmt.Errorf("point %d exceeded max retries (%d)", pointIndex, maxRetries)
}

// checkAlarm 检查采集数据是否触发报警，若触发则阻塞等待用户决策。
// 返回决策动作：continue/skip/recollect/stop。
func (s *Service) checkAlarm(ctx context.Context, pointIndex int, data []float64) (string, error) {
	s.mu.Lock()
	point := s.pressurePoints[pointIndex-1]
	alarmConfig := s.alarmConfig
	channels := s.config.Channels
	maxPressure := s.config.MaxPressure
	minPressure := s.config.MinPressure
	s.mu.Unlock()

	if len(data) == 0 {
		return workflow.AlarmDecisionContinue, nil
	}

	// 多通道报警判定
	channelData := make(map[int]float64)
	for i, ch := range channels {
		if i < len(data) {
			channelData[ch] = data[i]
		}
	}

	result := defaultAlarmService.EvaluateMultiChannel(alarmConfig, point.TargetPressure, maxPressure, minPressure, channelData)
	if !result.Triggered {
		return workflow.AlarmDecisionContinue, nil
	}

	// 触发报警，进入等待报警确认状态
	if err := s.sessionMachine.Transition(domain.SessionStateAwaitAlarmResolution); err != nil {
		// 若状态机不支持该迁移，记录日志后继续（不阻塞采集流程）
		log.Printf("[calibration] checkAlarm: transition to await_alarm_resolution failed: %v", err)
		return workflow.AlarmDecisionContinue, nil
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

		if err := defaultAlarmService.ValidateDecision(decision); err != nil {
			return "", err
		}

		s.publish(events.EventCalibrationAlarmResolved, map[string]any{
			"pointIndex": pointIndex,
			"decision":   decision,
			"triggered":  true,
		})

		switch decision {
		case workflow.AlarmDecisionStop:
			_ = s.sessionMachine.Transition(domain.SessionStateStopped)
			s.publishSessionState()
			return workflow.AlarmDecisionStop, nil
		case workflow.AlarmDecisionRecollect:
			_ = s.sessionMachine.Transition(domain.SessionStatePointDone)
			s.publishSessionState()
			return workflow.AlarmDecisionRecollect, nil
		case workflow.AlarmDecisionSkip:
			_ = s.sessionMachine.Transition(domain.SessionStatePointDone)
			s.publishSessionState()
			return workflow.AlarmDecisionSkip, nil
		default:
			_ = s.sessionMachine.Transition(domain.SessionStatePointDone)
			s.publishSessionState()
			return workflow.AlarmDecisionContinue, nil
		}
	case <-ctx.Done():
		s.alarmMu.Lock()
		s.alarmPending = false
		s.alarmCh = nil
		s.alarmMu.Unlock()
		return "", ctx.Err()
	}
}

// ResolveAlarm 用户确认报警，传入决策动作：continue/skip/recollect/stop。
func (s *Service) ResolveAlarm(decision string) error {
	if err := defaultAlarmService.ValidateDecision(decision); err != nil {
		return err
	}

	s.alarmMu.Lock()
	defer s.alarmMu.Unlock()

	if !s.alarmPending || s.alarmCh == nil {
		return ErrNoPendingAlarm
	}

	select {
	case s.alarmCh <- decision:
		// 事件由 checkAlarm 收到决策后统一发布，避免重复 SSE 推送。
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
	s.pressurePoints[pointIndex-1].Status = domain.PointStatusPending
	s.pressurePoints[pointIndex-1].CollectedData = nil
	s.pressurePoints[pointIndex-1].ActualPressure = 0
	s.currentPoint = pointIndex - 1
	controlMode := s.config.ControlMode
	hasPressureDriver := s.getPressureDriver() != nil
	s.mu.Unlock()

	s.publish(events.EventPointRetry, map[string]any{"pointIndex": pointIndex})

	// 手动模式且未连接打压设备时，仅重置为待确认，
	// 由操作者再次确认后执行采集，不自动触发打压链路。
	if controlMode == "manual" && !hasPressureDriver {
		return nil
	}

	return s.collectPoint(ctx, pointIndex)
}

// PauseAutoCollection 暂停自动采集，将状态机迁移到 paused。
func (s *Service) PauseAutoCollection() error {
	s.StopAutoCollection()

	if err := s.sessionMachine.Transition(domain.SessionStatePaused); err != nil {
		return fmt.Errorf("transition to paused: %w", err)
	}
	s.publishSessionState()
	return nil
}

// ResumeAutoCollection 恢复自动采集，重新启动从当前点的自动采集循环。
func (s *Service) ResumeAutoCollection(ctx context.Context) error {
	s.mu.Lock()
	s.currentPoint = s.resumePointIndexLocked()
	s.mu.Unlock()

	if err := s.sessionMachine.Transition(domain.SessionStatePressurizing); err != nil {
		return fmt.Errorf("transition to pressurizing: %w", err)
	}
	s.publishSessionState()
	return s.RunAutoCollection(ctx)
}

// StopAutoCollection 停止自动采集，取消后台 goroutine 并等待退出。
func (s *Service) StopAutoCollection() {
	s.autoCollectionMu.Lock()
	cancel := s.autoCollectionCancel
	s.autoCollectionCancel = nil
	s.autoCollectionCtx = nil
	s.autoCollectionMu.Unlock()

	if cancel != nil {
		cancel()
		s.autoCollectWg.Wait()
	}

	s.alarmMu.Lock()
	if s.alarmPending && s.alarmCh != nil {
		select {
		case s.alarmCh <- workflow.AlarmDecisionStop:
		default:
		}
		s.alarmPending = false
		s.alarmCh = nil
	}
	s.alarmMu.Unlock()
}

// IsAutoCollectionRunning 返回自动采集是否正在运行。
// ResolveStabilityTimeout 接收前端用户对稳定超时的决定。
// decision: "continue" 继续等待， "skip" 跳过当前点。
func (s *Service) ResolveStabilityTimeout(decision string) {
	select {
	case s.stabilityTimeoutCh <- decision:
	default:
	}
}

func (s *Service) IsAutoCollectionRunning() bool {
	s.autoCollectionMu.Lock()
	defer s.autoCollectionMu.Unlock()
	return s.autoCollectionCancel != nil
}

func (s *Service) markPointError(pointIndex int, reason string) {
	_ = reason
	s.updatePointStatus(pointIndex, domain.PointStatusError)
}

func (s *Service) publishSessionState() {
	s.publish(events.EventSessionStateChanged, map[string]any{
		"state": string(s.sessionMachine.State()),
	})
}

// updatePointStatus 更新压力点状态并发布统一事件。
func (s *Service) updatePointStatus(pointIndex int, status string) {
	s.mu.Lock()
	if pointIndex < 1 || pointIndex > len(s.pressurePoints) {
		s.mu.Unlock()
		return
	}
	point := &s.pressurePoints[pointIndex-1]
	point.Status = status
	pointSnapshot := *point
	if point.CollectedData != nil {
		pointSnapshot.CollectedData = append([]float64(nil), point.CollectedData...)
	}
	s.mu.Unlock()

	s.publish(events.EventCalibrationPointStatus, pointSnapshot)
}

// resumePointIndexLocked 计算恢复自动采集时的起始压力点下标（0-based）。
func (s *Service) resumePointIndexLocked() int {
	if len(s.pressurePoints) == 0 {
		return 0
	}

	start := s.currentPoint
	if start < 0 {
		start = 0
	}
	if start >= len(s.pressurePoints) {
		return len(s.pressurePoints)
	}

	for i := start; i < len(s.pressurePoints); i++ {
		if !isPointTerminalStatus(s.pressurePoints[i].Status) {
			return i
		}
	}

	for i := 0; i < start; i++ {
		if !isPointTerminalStatus(s.pressurePoints[i].Status) {
			return i
		}
	}

	return len(s.pressurePoints)
}

func isPointTerminalStatus(status string) bool {
	return status == domain.PointStatusCompleted || status == domain.PointStatusSkipped
}
