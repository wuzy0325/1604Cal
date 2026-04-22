package http

import (
	"fmt"
	"net/http"

	"cal1604/internal/domain"
	apperrors "cal1604/internal/errors"
)

type sessionStatePayload struct {
	State string `json:"state"`
}

func (s *apiServer) sessionStateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	writeSuccess(w, http.StatusOK, sessionStatePayload{State: string(s.sessionMachine.State())})
}

func (s *apiServer) sessionStartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// 启动前硬校验：阀门、设备、通道、配置等前置条件
	if err := s.calibrationService.ValidateStartPrerequisites(r.Context()); err != nil {
		writeError(w, fmt.Errorf("%w: %s", apperrors.ErrPrerequisiteNotMet, err.Error()))
		return
	}

	// 启动标定编排（含状态迁移、WTN1604 命令、自动采集入口）
	if err := s.calibrationService.StartCalibration(r.Context()); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, sessionStatePayload{State: string(s.sessionMachine.State())})
}

func (s *apiServer) sessionPauseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// 暂停自动采集
	if err := s.calibrationService.PauseAutoCollection(); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, sessionStatePayload{State: string(s.sessionMachine.State())})
}

func (s *apiServer) sessionResumeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// 恢复自动采集
	if err := s.calibrationService.ResumeAutoCollection(r.Context()); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, sessionStatePayload{State: string(s.sessionMachine.State())})
}

func (s *apiServer) sessionStopHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// 执行结束收尾：停压、停止采集循环、状态迁移到 stopped
	if err := s.calibrationService.EndCalibration(r.Context()); err != nil {
		writeError(w, err)
		return
	}

	// EndCalibration 不做状态迁移到 stopped，由 handler 层完成
	if err := s.sessionMachine.Transition(domain.SessionStateStopped); err != nil {
		// 尝试从 pressurizing -> paused -> stopped
		if s.sessionMachine.State() == domain.SessionStatePressurizing {
			_ = s.sessionMachine.Transition(domain.SessionStatePaused)
			s.publishSessionState()
			_ = s.sessionMachine.Transition(domain.SessionStateStopped)
		}
	}
	s.publishSessionState()

	writeSuccess(w, http.StatusOK, sessionStatePayload{State: string(s.sessionMachine.State())})
}

func (s *apiServer) publishSessionState() {
	publishEvent("session.state.changed", map[string]any{
		"state": string(s.sessionMachine.State()),
	})
}
