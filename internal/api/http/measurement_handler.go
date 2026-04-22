package http

import (
	"encoding/json"
	"net/http"

	apperrors "cal1604/internal/errors"
)

type measurementStartRequest struct {
	Channels []int `json:"channels"`
}

func (s *apiServer) measurementStateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"state": string(s.measurementService.State())})
}

func (s *apiServer) measurementStartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req measurementStartRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	if len(req.Channels) == 0 {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	if err := s.measurementService.Start(r.Context(), req.Channels); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"state": "collecting"})
}

func (s *apiServer) measurementPauseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := s.measurementService.Pause(); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"state": "paused"})
}

func (s *apiServer) measurementStopHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := s.measurementService.Stop(); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"state": "idle"})
}

func (s *apiServer) measurementDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	rows, total := s.measurementService.GetData()
	writeSuccess(w, http.StatusOK, map[string]any{"rows": rows, "total": total})
}

func (s *apiServer) measurementExportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=measurement_data.csv")

	if err := s.measurementService.WriteCSV(w); err != nil {
		writeError(w, err)
		return
	}
}
