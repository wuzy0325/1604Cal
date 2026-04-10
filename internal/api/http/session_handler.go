package http

import (
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

	current := s.sessionMachine.State()
	if current == domain.SessionStateIdle {
		if err := s.sessionMachine.Transition(domain.SessionStateReady); err != nil {
			writeError(w, apperrors.ErrInvalidStateTransition)
			return
		}
	}

	if err := s.sessionMachine.Transition(domain.SessionStatePressurizing); err != nil {
		writeError(w, apperrors.ErrInvalidStateTransition)
		return
	}

	writeSuccess(w, http.StatusOK, sessionStatePayload{State: string(s.sessionMachine.State())})
}

func (s *apiServer) sessionPauseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := s.sessionMachine.Transition(domain.SessionStatePaused); err != nil {
		writeError(w, apperrors.ErrInvalidStateTransition)
		return
	}

	writeSuccess(w, http.StatusOK, sessionStatePayload{State: string(s.sessionMachine.State())})
}

func (s *apiServer) sessionResumeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := s.sessionMachine.Transition(domain.SessionStatePressurizing); err != nil {
		writeError(w, apperrors.ErrInvalidStateTransition)
		return
	}

	writeSuccess(w, http.StatusOK, sessionStatePayload{State: string(s.sessionMachine.State())})
}

func (s *apiServer) sessionStopHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := s.sessionMachine.Transition(domain.SessionStateStopped); err != nil {
		if s.sessionMachine.State() == domain.SessionStatePressurizing {
			if pauseErr := s.sessionMachine.Transition(domain.SessionStatePaused); pauseErr != nil {
				writeError(w, apperrors.ErrInvalidStateTransition)
				return
			}

			if stopErr := s.sessionMachine.Transition(domain.SessionStateStopped); stopErr != nil {
				writeError(w, apperrors.ErrInvalidStateTransition)
				return
			}
		} else {
			writeError(w, apperrors.ErrInvalidStateTransition)
			return
		}
	}

	writeSuccess(w, http.StatusOK, sessionStatePayload{State: string(s.sessionMachine.State())})
}
