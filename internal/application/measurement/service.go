package measurement

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"sync"
	"time"

	"cal1604/internal/application/session"
)

// State 计量模块状态。
type State string

const (
	StateIdle       State = "idle"
	StateCollecting State = "collecting"
	StatePaused     State = "paused"
)

// 获取状态迁移表。
var transitions = map[State]map[State]struct{}{
	StateIdle:       {StateCollecting: {}},
	StateCollecting: {StatePaused: {}, StateIdle: {}},
	StatePaused:     {StateCollecting: {}, StateIdle: {}},
}

// CollectedRow 单次采集的数据行。
type CollectedRow struct {
	Timestamp string             `json:"timestamp"`
	Channels  map[string]float64 `json:"channels"`
}

// EventPublisher 广播事件。
type EventPublisher func(eventType string, data any)

// Service 计量模块服务，管理简化的采集工作流。
type Service struct {
	mu      sync.Mutex
	state   State
	sess    *session.Service
	publish EventPublisher

	channels []int
	rows     []CollectedRow

	// collectCtx 用于控制采集 goroutine 的生命周期。
	collectCtx    context.Context
	collectCancel context.CancelFunc
	collectMu     sync.Mutex
}

// NewService 创建计量服务。
func NewService(sess *session.Service, publisher EventPublisher) *Service {
	if publisher == nil {
		publisher = func(string, any) {}
	}
	return &Service{
		state:   StateIdle,
		sess:    sess,
		publish: publisher,
	}
}

// State 返回当前计量状态。
func (s *Service) State() State {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state
}

// Start 启动计量采集。
func (s *Service) Start(ctx context.Context, channels []int) error {
	s.mu.Lock()
	if err := s.canTransition(StateCollecting); err != nil {
		s.mu.Unlock()
		return fmt.Errorf("start measurement: %w", err)
	}

	// 检查计量设备是否已绑定
	if s.sess.MeasureDriver() == nil {
		s.mu.Unlock()
		return session.ErrMeasureDeviceNotSet
	}

	s.channels = channels
	s.rows = nil
	s.state = StateCollecting
	s.mu.Unlock()

	s.publish("measurement.state_changed", map[string]any{"state": string(StateCollecting)})

	// 启动后台采集循环
	s.startCollectLoop(ctx)

	return nil
}

// Pause 暂停计量采集。
func (s *Service) Pause() error {
	s.mu.Lock()
	if err := s.canTransition(StatePaused); err != nil {
		s.mu.Unlock()
		return fmt.Errorf("pause measurement: %w", err)
	}
	s.state = StatePaused
	s.mu.Unlock()

	s.stopCollectLoop()
	s.publish("measurement.state_changed", map[string]any{"state": string(StatePaused)})

	return nil
}

// Stop 停止计量采集，重置为 idle。
func (s *Service) Stop() error {
	s.mu.Lock()
	if s.state == StateIdle {
		s.mu.Unlock()
		return fmt.Errorf("not running")
	}
	s.state = StateIdle
	s.mu.Unlock()

	s.stopCollectLoop()
	s.publish("measurement.state_changed", map[string]any{"state": string(StateIdle)})

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

// WriteCSV 将已采集数据写入 CSV 格式。
func (s *Service) WriteCSV(w io.Writer) error {
	s.mu.Lock()
	rows := make([]CollectedRow, len(s.rows))
	copy(rows, s.rows)
	channels := make([]int, len(s.channels))
	copy(channels, s.channels)
	s.mu.Unlock()

	if len(rows) == 0 {
		return fmt.Errorf("no data to export")
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

// startCollectLoop 启动后台采集 goroutine。
func (s *Service) startCollectLoop(ctx context.Context) {
	s.collectMu.Lock()
	if s.collectCancel != nil {
		s.collectMu.Unlock()
		return
	}
	s.collectCtx, s.collectCancel = context.WithCancel(ctx)
	collectCtx := s.collectCtx
	s.collectMu.Unlock()

	go func() {
		defer func() {
			s.collectMu.Lock()
			s.collectCancel = nil
			s.collectCtx = nil
			s.collectMu.Unlock()
		}()

		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-collectCtx.Done():
				return
			case <-ticker.C:
				s.mu.Lock()
				if s.state != StateCollecting {
					s.mu.Unlock()
					return
				}
				channels := s.channels
				s.mu.Unlock()

				data, err := s.sess.ReadMeasureData(collectCtx)
				if err != nil {
					continue
				}

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

				s.publish("measurement.data_updated", map[string]any{
					"timestamp": row.Timestamp,
					"channels":  chMap,
				})
			}
		}
	}()
}

// stopCollectLoop 停止后台采集 goroutine。
func (s *Service) stopCollectLoop() {
	s.collectMu.Lock()
	if s.collectCancel != nil {
		s.collectCancel()
		s.collectCancel = nil
		s.collectCtx = nil
	}
	s.collectMu.Unlock()
}

// canTransition 检查从当前状态是否可迁移到目标状态。
func (s *Service) canTransition(target State) error {
	allowed, ok := transitions[s.state]
	if !ok {
		return fmt.Errorf("state %s has no transitions", s.state)
	}
	if _, exists := allowed[target]; !exists {
		return fmt.Errorf("invalid transition: %s -> %s", s.state, target)
	}
	return nil
}
