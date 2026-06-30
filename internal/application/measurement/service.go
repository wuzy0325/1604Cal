package measurement

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"cal1604/internal/application/session"
	"cal1604/internal/domain"
	apperrors "cal1604/internal/errors"
	"cal1604/internal/events"
	"cal1604/internal/infrastructure/driver"
	"cal1604/internal/workflow"
)

var ErrPointSkipped = errors.New("point skipped by user")

// ErrRecollectPoint 表示用户选择重新采集当前压力点，调用方应重试当前点。
var ErrRecollectPoint = errors.New("point recollect requested by user")

// allChannels 全部16个通道，用于始终读取全部通道数据。
var allChannels = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

// CollectedRow 单次采集的数据行。
type CollectedRow struct {
	Timestamp string             `json:"timestamp"`
	Channels  map[string]float64 `json:"channels"`
}

// EventPublisher 广播事件。
type EventPublisher func(eventType string, data any)

// StartPrerequisiteConfig 定义计量启动门禁配置。
// 与标定模块同款：阀门=校准模式是必要条件。
type StartPrerequisiteConfig struct {
	EnforceValveCalibration bool
}

// defaultStartPrerequisiteConfig 返回默认启动门禁配置。
// 默认开启：阀门=校准模式是标定与计量启动的必要条件。
func defaultStartPrerequisiteConfig() StartPrerequisiteConfig {
	return StartPrerequisiteConfig{EnforceValveCalibration: true}
}

// Service 计量模块服务，管理简化的采集工作流。
type Service struct {
	mu          sync.Mutex
	coordinator *workflow.WorkflowCoordinator
	sess        *session.Service
	publish     EventPublisher

	config  domain.WorkflowConfig
	points  []domain.PressurePoint
	session *Session

	channels []int
	rows     []CollectedRow

	measureDeviceID  string
	pressureDeviceID string

	alarmConfig  domain.AlarmConfig
	alarmPending bool
	currentAlarm *Alarm

	// alarmCh 用于在自动采集过程中阻塞等待用户确认报警。
	alarmCh chan string

	// collectCtx 用于控制采集 goroutine 的生命周期。
	collectCtx    context.Context
	collectCancel context.CancelFunc
	collectMu     sync.Mutex
	collectWg     sync.WaitGroup

	// autoCollectCtx 用于控制自动按点采集 goroutine 的生命周期。
	autoCollectCtx    context.Context
	autoCollectCancel context.CancelFunc
	autoCollectMu     sync.Mutex
	autoCollectWg     sync.WaitGroup

	// stabilityTimeoutCh 用于等待前端用户对稳定超时的决定。
	stabilityTimeoutCh chan string

	// sessionStore 历史会话持久化存储。
	sessionStore *SessionStore

	// startPrerequisiteConfig 启动门禁配置（阀门=校准必要条件等）。
	startPrerequisiteConfig StartPrerequisiteConfig
}

// NewService 创建计量服务。
func NewService(sess *session.Service, publisher EventPublisher, coordinator *workflow.WorkflowCoordinator) *Service {
	if publisher == nil {
		publisher = func(string, any) {}
	}
	return &Service{
		coordinator:             coordinator,
		sess:                    sess,
		publish:                 publisher,
		stabilityTimeoutCh:      make(chan string, 1),
		startPrerequisiteConfig: defaultStartPrerequisiteConfig(),
	}
}

// SetStartPrerequisiteConfig 设置计量启动门禁配置。
func (s *Service) SetStartPrerequisiteConfig(cfg StartPrerequisiteConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.startPrerequisiteConfig = cfg
}

// SetSessionStore 设置会话持久化存储。
func (s *Service) SetSessionStore(store *SessionStore) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessionStore = store
}

// State 返回当前计量状态。
func (s *Service) State() domain.SessionState {
	return s.coordinator.State()
}

// Start 启动计量采集。
func (s *Service) Start(ctx context.Context, channels []int) error {
	s.mu.Lock()
	// 检查计量设备是否已绑定。
	if s.sess.MeasureDriver() == nil {
		s.mu.Unlock()
		return session.ErrMeasureDeviceNotSet
	}
	enforceValveGate := s.startPrerequisiteConfig.EnforceValveCalibration
	measureDriver := s.sess.MeasureDriver()
	s.mu.Unlock()

	// 阀门门禁：阀门=校准模式是计量启动的必要条件。
	// 必须在 coordinator.Begin 之前校验，避免占用工作流锁后又因门禁失败回滚。
	// 锁外读阀（最长 3s I/O），不阻塞会话锁。
	if enforceValveGate {
		valveStatus, err := measureDriver.ReadValveStatus(ctx)
		if err != nil {
			return fmt.Errorf("%w: read valve status: %v", apperrors.ErrPrerequisiteNotMet, err)
		}
		if domain.ValveState(valveStatus) != domain.ValveStateCalibration {
			return fmt.Errorf("%w: valve must be in calibration state, current: %s",
				apperrors.ErrPrerequisiteNotMet, valveStatus)
		}
	}

	// 单活工作流冲突校验
	if err := s.coordinator.Begin(workflow.OwnerMeasurement); err != nil {
		return err
	}

	s.mu.Lock()
	stateChanges := make([]domain.SessionState, 0, 3)
	currentState := s.coordinator.State()

	// 从暂停恢复时保留已采集数据；其他入口重置数据。
	if currentState != domain.SessionStatePaused {
		s.rows = nil
	}
	// channels 仅用于报警通道配置，不用于限制数据采集（始终全采16通道）
	s.channels = append([]int(nil), channels...)

	switch currentState {
	case domain.SessionStateIdle, domain.SessionStateCompleted, domain.SessionStateError, domain.SessionStateReady:
		if err := s.coordinator.Machine().Transition(domain.SessionStateCollecting); err != nil {
			s.mu.Unlock()
			return fmt.Errorf("start measurement: %w", err)
		}
		stateChanges = append(stateChanges, domain.SessionStateCollecting)
	case domain.SessionStatePaused:
		if err := s.coordinator.Machine().Transition(domain.SessionStateCollecting); err != nil {
			s.mu.Unlock()
			return fmt.Errorf("resume measurement: %w", err)
		}
		stateChanges = append(stateChanges, domain.SessionStateCollecting)
	default:
		err := fmt.Errorf("invalid transition: %s -> %s", currentState, domain.SessionStateCollecting)
		s.mu.Unlock()
		return fmt.Errorf("start measurement: %w", err)
	}
	s.syncSessionStatusLocked(s.coordinator.State())

	s.mu.Unlock()

	for _, state := range stateChanges {
		s.publish(events.EventMeasurementStateChanged, map[string]any{"state": string(state)})
	}

	// 启动后台采集循环
	s.startCollectLoop(ctx)

	return nil
}

// Pause 暂停计量采集。
func (s *Service) Pause() error {
	s.mu.Lock()
	if err := s.coordinator.Machine().Transition(domain.SessionStatePaused); err != nil {
		s.mu.Unlock()
		return fmt.Errorf("pause measurement: %w", err)
	}
	s.syncSessionStatusLocked(domain.SessionStatePaused)
	s.mu.Unlock()

	s.stopCollectLoop()
	s.StopAutoCollect()
	s.publish(events.EventMeasurementStateChanged, map[string]any{"state": string(domain.SessionStatePaused)})

	return nil
}

// Stop 停止计量采集，重置为 idle。
func (s *Service) Stop() error {
	s.mu.Lock()
	if s.coordinator.State() == domain.SessionStateIdle {
		s.mu.Unlock()
		return fmt.Errorf("not running")
	}

	stateChanges := make([]domain.SessionState, 0, 2)
	switch s.coordinator.State() {
	case domain.SessionStateCollecting:
		if err := s.coordinator.Machine().Transition(domain.SessionStateCompleted); err != nil {
			s.mu.Unlock()
			return fmt.Errorf("stop measurement: %w", err)
		}
		stateChanges = append(stateChanges, domain.SessionStateCompleted)
		if err := s.coordinator.Machine().Transition(domain.SessionStateIdle); err != nil {
			s.mu.Unlock()
			return fmt.Errorf("stop measurement: %w", err)
		}
		stateChanges = append(stateChanges, domain.SessionStateIdle)
	case domain.SessionStatePressurizing, domain.SessionStateStabilizing:
		if err := s.coordinator.Machine().Transition(domain.SessionStatePaused); err != nil {
			s.mu.Unlock()
			return fmt.Errorf("stop measurement: %w", err)
		}
		stateChanges = append(stateChanges, domain.SessionStatePaused)
		if err := s.coordinator.Machine().Transition(domain.SessionStateIdle); err != nil {
			s.mu.Unlock()
			return fmt.Errorf("stop measurement: %w", err)
		}
		stateChanges = append(stateChanges, domain.SessionStateIdle)
	case domain.SessionStateReady, domain.SessionStatePaused, domain.SessionStateCompleted, domain.SessionStateError:
		if err := s.coordinator.Machine().Transition(domain.SessionStateIdle); err != nil {
			s.mu.Unlock()
			return fmt.Errorf("stop measurement: %w", err)
		}
		stateChanges = append(stateChanges, domain.SessionStateIdle)
	default:
		s.mu.Unlock()
		return fmt.Errorf("stop measurement: unsupported state %s", s.coordinator.State())
	}
	now := time.Now()
	s.finishSessionLocked(domain.SessionStateStopped, &now)

	// 清理报警阻塞等待
	if s.alarmPending && s.alarmCh != nil {
		select {
		case s.alarmCh <- workflow.AlarmDecisionStop:
		default:
		}
		s.alarmPending = false
		s.alarmCh = nil
	}

	s.mu.Unlock()

	s.stopCollectLoop()
	s.StopAutoCollect()
	for _, state := range stateChanges {
		s.publish(events.EventMeasurementStateChanged, map[string]any{"state": string(state)})
	}

	s.coordinator.End()

	return nil
}

// SetState 进行显式状态切换，并发布状态变化事件。
func (s *Service) SetState(state domain.SessionState) error {
	s.mu.Lock()
	if err := s.coordinator.Machine().Transition(state); err != nil {
		s.mu.Unlock()
		return err
	}
	s.syncSessionStatusLocked(state)
	s.mu.Unlock()

	s.publish(events.EventMeasurementStateChanged, map[string]any{"state": string(state)})
	return nil
}

// GetData 返回已采集的数据。
func (s *Service) GetData() ([]CollectedRow, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]CollectedRow, len(s.rows))
	copy(result, s.rows)
	return result, len(result)
}

// ListSessions 返回历史会话列表。
func (s *Service) ListSessions() ([]*Session, error) {
	s.mu.Lock()
	store := s.sessionStore
	s.mu.Unlock()
	if store == nil {
		return nil, nil
	}
	return store.List()
}

// GetSessionByID 根据 ID 加载历史会话。
func (s *Service) GetSessionByID(id string) (*Session, error) {
	s.mu.Lock()
	store := s.sessionStore
	s.mu.Unlock()
	if store == nil {
		return nil, fmt.Errorf("session store not configured")
	}
	return store.Get(id)
}

// WriteCSV 将已采集数据写入 CSV 格式，始终输出全部16通道。
func (s *Service) WriteCSV(w io.Writer) error {
	s.mu.Lock()
	rows := make([]CollectedRow, len(s.rows))
	copy(rows, s.rows)
	points := make([]domain.PressurePoint, len(s.points))
	copy(points, s.points)
	s.mu.Unlock()

	// 优先使用实时采集行数据，否则从按点采集的压力点数据生成
	if len(rows) == 0 {
		rows = rowsFromPoints(points)
	}

	if len(rows) == 0 {
		return apperrors.ErrNoData
	}

	cw := csv.NewWriter(w)
	defer cw.Flush()

	// 表头（始终输出全部16通道）
	header := []string{"timestamp"}
	for _, ch := range allChannels {
		header = append(header, fmt.Sprintf("channel_%d", ch))
	}
	if err := cw.Write(header); err != nil {
		return err
	}

	// 数据行
	for _, row := range rows {
		record := []string{row.Timestamp}
		for _, ch := range allChannels {
			key := fmt.Sprintf("%d", ch)
			if v, ok := row.Channels[key]; ok {
				record = append(record, fmt.Sprintf("%.4f", v))
			} else {
				record = append(record, "")
			}
		}
		if err := cw.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// rowsFromPoints 从按点采集的压力点数据生成 CollectedRow，用于标定模式导出。
func rowsFromPoints(points []domain.PressurePoint) []CollectedRow {
	var result []CollectedRow
	for _, p := range points {
		if len(p.CollectedData) == 0 || p.Status != "completed" {
			continue
		}
		chMap := make(map[string]float64, len(p.CollectedData))
		for i, v := range p.CollectedData {
			chMap[fmt.Sprintf("%d", i+1)] = v
		}
		result = append(result, CollectedRow{
			Timestamp: p.CollectTime,
			Channels:  chMap,
		})
	}
	return result
}

// startCollectLoop 启动后台采集 goroutine。
// 使用 context.Background() 确保采集循环不受调用方 context 生命周期影响，
// 仅通过 stopCollectLoop() 控制停止。
func (s *Service) startCollectLoop(_ context.Context) {
	s.collectMu.Lock()
	if s.collectCancel != nil {
		s.collectMu.Unlock()
		return
	}
	s.collectCtx, s.collectCancel = context.WithCancel(context.Background())
	collectCtx := s.collectCtx
	s.collectWg.Add(1)
	s.collectMu.Unlock()

	const maxConsecutiveErrors = 10

	go func() {
		defer s.collectWg.Done()

		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		consecutiveErrors := 0

		for {
			select {
			case <-collectCtx.Done():
				return
			case <-ticker.C:
				s.mu.Lock()
				if s.coordinator.State() != domain.SessionStateCollecting {
					s.mu.Unlock()
					return
				}
				s.mu.Unlock()

				pollCtx := driver.WithPollContext(collectCtx)
				data, err := s.sess.ReadMeasureData(pollCtx, s.sess.Token())
				if err != nil {
					consecutiveErrors++
					if consecutiveErrors >= maxConsecutiveErrors {
						s.publish(events.EventMeasurementStateChanged, map[string]any{
							"state": string(domain.SessionStateError),
							"error": fmt.Sprintf("连续%d次采集失败: %v", consecutiveErrors, err),
						})
						s.mu.Lock()
						_ = s.coordinator.Machine().Transition(domain.SessionStateError)
						s.mu.Unlock()
						return
					}
					continue
				}
				consecutiveErrors = 0

				// 构建通道映射（始终包含全部16通道）
				chMap := make(map[string]float64, 16)
				for i, ch := range allChannels {
					if i < len(data) {
						chMap[fmt.Sprintf("%d", ch)] = data[i]
					}
				}

				row := CollectedRow{
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					Channels:  chMap,
				}

				s.mu.Lock()
				s.rows = append(s.rows, row)
				s.mu.Unlock()

				s.publish(events.EventMeasurementDataUpdated, map[string]any{
					"timestamp": row.Timestamp,
					"channels":  chMap,
				})
			}
		}
	}()
}

// stopCollectLoop 停止后台采集 goroutine 并等待其退出。
func (s *Service) stopCollectLoop() {
	s.collectMu.Lock()
	cancel := s.collectCancel
	s.collectCancel = nil
	s.collectCtx = nil
	s.collectMu.Unlock()

	if cancel != nil {
		cancel()
		s.collectWg.Wait()
	}
}

// StartAutoCollect 启动自动按点采集 goroutine，返回可取消的 context。
// 调用方负责在 Stop/Pause 时通过 StopAutoCollect 取消。
func (s *Service) StartAutoCollect() {
	s.autoCollectMu.Lock()
	defer s.autoCollectMu.Unlock()

	if s.autoCollectCancel != nil {
		log.Printf("[measurement] StartAutoCollect skipped: already running")
		return // 已经在运行
	}

	s.autoCollectCtx, s.autoCollectCancel = context.WithCancel(context.Background())
	ctx := s.autoCollectCtx
	s.autoCollectWg.Add(1)

	go func() {
		defer s.autoCollectWg.Done()
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[measurement] auto collection PANIC: %v", r)
				s.SetState(domain.SessionStateError)
			}
		}()
		defer func() {
			s.autoCollectMu.Lock()
			s.autoCollectCancel = nil
			s.autoCollectCtx = nil
			s.autoCollectMu.Unlock()
		}()
		if err := s.RunAutoCollection(ctx); err != nil {
			log.Printf("[measurement] auto collection failed: %v", err)
			s.SetState(domain.SessionStateError)
		}
	}()
}

// StopAutoCollect 停止自动按点采集 goroutine 并等待其退出。
func (s *Service) StopAutoCollect() {
	s.autoCollectMu.Lock()
	cancel := s.autoCollectCancel
	s.autoCollectCancel = nil
	s.autoCollectCtx = nil
	s.autoCollectMu.Unlock()

	if cancel != nil {
		cancel()
		s.autoCollectWg.Wait()
	}
}

func (s *Service) syncSessionStatusLocked(state domain.SessionState) {
	if s.session == nil {
		return
	}
	s.session.Status = state
}

func (s *Service) finishSessionLocked(state domain.SessionState, endTime *time.Time) {
	if s.session == nil {
		return
	}
	s.session.Status = state
	s.session.EndTime = endTime

	if s.sessionStore != nil {
		snapshot := *s.session
		go func() {
			_ = s.sessionStore.Save(&snapshot)
		}()
	}
}
