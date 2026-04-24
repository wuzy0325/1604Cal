package measurement

import (
	"context"
	"fmt"
	"time"

	"cal1604/internal/application/session"
	apperrors "cal1604/internal/errors"
)

// Session 表示 measurement 自己的流程会话。
type Session struct {
	ID               string     `json:"id"`
	StartTime        time.Time  `json:"startTime"`
	EndTime          *time.Time `json:"endTime,omitempty"`
	Config           Config     `json:"config"`
	Points           []Point    `json:"points"`
	MeasureDeviceID  string     `json:"measureDeviceId"`
	PressureDeviceID string     `json:"pressureDeviceId"`
	Status           State      `json:"status"`
}

// StartWorkflow 启动 measurement 自己的业务流程会话。
// 当前阶段仅完成“参数 + 点位计划 -> ready 会话”的收口，
// 自动/手动采集编排在后续任务继续补齐。
func (s *Service) StartWorkflow(_ context.Context, channels []int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.sess.MeasureDriver() == nil {
		return session.ErrMeasureDeviceNotSet
	}

	if len(channels) == 0 {
		return fmt.Errorf("%w: no measurement channels selected", apperrors.ErrInvalidArgument)
	}

	if len(s.points) == 0 {
		return fmt.Errorf("%w: measurement points not generated", apperrors.ErrInvalidArgument)
	}

	s.config.Channels = append([]int(nil), channels...)
	s.channels = append([]int(nil), channels...)
	s.rows = nil
	s.sess.SetChannels(channels)

	if err := s.setStateLocked(StateReady); err != nil {
		return fmt.Errorf("start measurement workflow: %w", err)
	}

	s.session = &Session{
		ID:               fmt.Sprintf("measurement-%d", time.Now().UnixMilli()),
		StartTime:        time.Now(),
		Config:           s.config,
		Points:           append([]Point(nil), s.points...),
		MeasureDeviceID:  s.sess.MeasureDeviceID(),
		PressureDeviceID: s.sess.PressureDeviceID(),
		Status:           s.state,
	}

	s.publish("measurement.state_changed", map[string]any{"state": string(StateReady)})
	return nil
}

// GetSession 返回当前 measurement 流程会话快照。
func (s *Service) GetSession() *Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.session == nil {
		return nil
	}

	cloned := *s.session
	cloned.Points = make([]Point, len(s.session.Points))
	for i, point := range s.session.Points {
		cloned.Points[i] = point
		if point.CollectedData != nil {
			cloned.Points[i].CollectedData = append([]float64(nil), point.CollectedData...)
		}
	}
	if len(cloned.Config.Channels) > 0 {
		cloned.Config.Channels = append([]int(nil), cloned.Config.Channels...)
	}
	return &cloned
}
