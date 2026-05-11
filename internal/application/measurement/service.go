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
	"cal1604/internal/workflow"
)

var ErrPointSkipped = errors.New("point skipped by user")

// CollectedRow 单次采集的数据行。
type CollectedRow struct {
	Timestamp string             `json:"timestamp"`
	Channels  map[string]float64 `json:"channels"`
}

// EventPublisher 广播事件。
type EventPublisher func(eventType string, data any)

// Service 计量模块服务，管理简化的采集工作流。
type Service struct {
	mu             sync.Mutex
	sessionMachine *workflow.SessionMachine
	sess           *session.Service
	publish        EventPublisher

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
}

// NewService 创建计量服务。
func NewService(sess *session.Service, publisher EventPublisher) *Service {
	if publisher == nil {
		publisher = func(string, any) {}
	}
	return &Service{
		sessionMachine:     workflow.NewSessionMachine(),
		sess:               sess,
		publish:            publisher,
		stabilityTimeoutCh: make(chan string, 1),
	}
}

// SetSessionStore 设置会话持久化存储。
func (s *Service) SetSessionStore(store *SessionStore) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessionStore = store
}

// State 返回当前计量状态。
func (s *Service) State() domain.SessionState {
	return s.sessionMachine.State()
}

// Start 启动计量采集。
func (s *Service) Start(ctx context.Context, channels []int) error {
	s.mu.Lock()
	// 检查计量设备是否已绑定。
	if s.sess.MeasureDriver() == nil {
		s.mu.Unlock()
		return session.ErrMeasureDeviceNotSet
	}

	stateChanges := make([]domain.SessionState, 0, 3)
	currentState := s.sessionMachine.State()

	// 从暂停恢复时保留已采集数据；其他入口重置数据。
	if currentState != domain.SessionStatePaused {
		s.rows = nil
	}
	s.channels = append([]int(nil), channels...)
	s.sess.SetChannels(channels)

	switch currentState {
	case domain.SessionStateIdle, domain.SessionStateCompleted, domain.SessionStateError, domain.SessionStateReady:
		if err := s.sessionMachine.Transition(domain.SessionStateCollecting); err != nil {
			s.mu.Unlock()
			return fmt.Errorf("start measurement: %w", err)
		}
		stateChanges = append(stateChanges, domain.SessionStateCollecting)
	case domain.SessionStatePaused:
		if err := s.sessionMachine.Transition(domain.SessionStateCollecting); err != nil {
			s.mu.Unlock()
			return fmt.Errorf("resume measurement: %w", err)
		}
		stateChanges = append(stateChanges, domain.SessionStateCollecting)
	default:
		err := fmt.Errorf("invalid transition: %s -> %s", currentState, domain.SessionStateCollecting)
		s.mu.Unlock()
		return fmt.Errorf("start measurement: %w", err)
	}
	s.syncSessionStatusLocked(s.sessionMachine.State())

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
	if err := s.sessionMachine.Transition(domain.SessionStatePaused); err != nil {
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
	if s.sessionMachine.State() == domain.SessionStateIdle {
		s.mu.Unlock()
		return fmt.Errorf("not running")
	}

	stateChanges := make([]domain.SessionState, 0, 2)
	switch s.sessionMachine.State() {
	case domain.SessionStateCollecting:
		if err := s.sessionMachine.Transition(domain.SessionStateCompleted); err != nil {
			s.mu.Unlock()
			return fmt.Errorf("stop measurement: %w", err)
		}
		stateChanges = append(stateChanges, domain.SessionStateCompleted)
		if err := s.sessionMachine.Transition(domain.SessionStateIdle); err != nil {
			s.mu.Unlock()
			return fmt.Errorf("stop measurement: %w", err)
		}
		stateChanges = append(stateChanges, domain.SessionStateIdle)
	case domain.SessionStatePressurizing, domain.SessionStateStabilizing:
		if err := s.sessionMachine.Transition(domain.SessionStatePaused); err != nil {
			s.mu.Unlock()
			return fmt.Errorf("stop measurement: %w", err)
		}
		stateChanges = append(stateChanges, domain.SessionStatePaused)
		if err := s.sessionMachine.Transition(domain.SessionStateIdle); err != nil {
			s.mu.Unlock()
			return fmt.Errorf("stop measurement: %w", err)
		}
		stateChanges = append(stateChanges, domain.SessionStateIdle)
	case domain.SessionStateReady, domain.SessionStatePaused, domain.SessionStateCompleted, domain.SessionStateError:
		if err := s.sessionMachine.Transition(domain.SessionStateIdle); err != nil {
			s.mu.Unlock()
			return fmt.Errorf("stop measurement: %w", err)
		}
		stateChanges = append(stateChanges, domain.SessionStateIdle)
	default:
		s.mu.Unlock()
		return fmt.Errorf("stop measurement: unsupported state %s", s.sessionMachine.State())
	}
	now := time.Now()
	s.finishSessionLocked(domain.SessionStateStopped, &now)

	s.mu.Unlock()

	s.stopCollectLoop()
	s.StopAutoCollect()
	for _, state := range stateChanges {
		s.publish(events.EventMeasurementStateChanged, map[string]any{"state": string(state)})
	}

	return nil
}

// SetState 进行显式状态切换，并发布状态变化事件。
func (s *Service) SetState(state domain.SessionState) error {
	s.mu.Lock()
	if err := s.sessionMachine.Transition(state); err != nil {
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

// WriteCSV 将已采集数据写入 CSV 格式。
func (s *Service) WriteCSV(w io.Writer) error {
	s.mu.Lock()
	rows := make([]CollectedRow, len(s.rows))
	copy(rows, s.rows)
	channels := make([]int, len(s.channels))
	copy(channels, s.channels)
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

	// 若 channels 为空，从 points 推断通道数
	if len(channels) == 0 {
		for _, p := range points {
			if len(p.CollectedData) > len(channels) {
				channels = make([]int, len(p.CollectedData))
				for i := range channels {
					channels[i] = i + 1
				}
			}
		}
	}

	cw := csv.NewWriter(w)
	defer cw.Flush()

	// 表头
	header := []string{"timestamp"}
	for _, ch := range channels {
		header = append(header, fmt.Sprintf("channel_%d", ch))
	}
	if err := cw.Write(header); err != nil {
		return err
	}

	// 数据行
	for _, row := range rows {
		record := []string{row.Timestamp}
		for _, ch := range channels {
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
				if s.sessionMachine.State() != domain.SessionStateCollecting {
					s.mu.Unlock()
					return
				}
				channels := s.channels
				s.mu.Unlock()

				data, err := s.sess.ReadMeasureData(collectCtx)
				if err != nil {
					consecutiveErrors++
					if consecutiveErrors >= maxConsecutiveErrors {
						s.publish(events.EventMeasurementStateChanged, map[string]any{
							"state": string(domain.SessionStateError),
							"error": fmt.Sprintf("连续%d次采集失败: %v", consecutiveErrors, err),
						})
						s.mu.Lock()
						_ = s.sessionMachine.Transition(domain.SessionStateError)
						s.mu.Unlock()
						return
					}
					continue
				}
				consecutiveErrors = 0

				// 构建通道映射
				chMap := make(map[string]float64, len(channels))
				for i, ch := range channels {
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
			// goroutine 退出时清理 cancel 标记，确保下次 StartAutoCollect 可重新启动
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
