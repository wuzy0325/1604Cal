package http

import (
	"encoding/json"
	"net/http"

	"cal1604/internal/application/measurement"
	apperrors "cal1604/internal/errors"
)

type measurementStartRequest struct {
	Channels []int `json:"channels"`
}

func (s *apiServer) measurementGeneratePointsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	points, err := s.measurementService.GeneratePressurePoints()
	if err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, points)
}

func (s *apiServer) measurementPointsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	writeSuccess(w, http.StatusOK, s.measurementService.GetPoints())
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

	if err := s.measurementService.StartWorkflow(r.Context(), req.Channels); err != nil {
		writeError(w, err)
		return
	}

	if err := s.measurementService.Start(r.Context(), req.Channels); err != nil {
		// Start 失败时回滚到 idle。
		_ = s.measurementService.Stop()
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"state": string(s.measurementService.State())})
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

func (s *apiServer) measurementAlarmConfigHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg := s.measurementService.GetAlarmConfig()
		writeSuccess(w, http.StatusOK, cfg)
	case http.MethodPost:
		var cfg measurement.AlarmConfig
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&cfg); err != nil {
			writeError(w, apperrors.ErrInvalidArgument)
			return
		}
		s.measurementService.SetAlarmConfig(cfg)
		writeSuccess(w, http.StatusOK, map[string]string{"status": "saved"})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *apiServer) measurementAlarmResolveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Decision string `json:"decision"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	if err := s.measurementService.ResolveAlarm(req.Decision); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"status": "resolved"})
}

func (s *apiServer) measurementAlarmPendingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]bool{"pending": s.measurementService.IsAlarmPending()})
}

func (s *apiServer) measurementAutoCollectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := s.measurementService.RunAutoCollection(r.Context()); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"state": string(s.measurementService.State())})
}

func (s *apiServer) measurementManualPressurizeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PointIndex int `json:"pointIndex"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	if err := s.measurementService.ManualPressurize(r.Context(), req.PointIndex); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"state": string(s.measurementService.State())})
}

func (s *apiServer) measurementManualCollectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PointIndex int `json:"pointIndex"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	if err := s.measurementService.ManualCollect(r.Context(), req.PointIndex); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"state": string(s.measurementService.State())})
}
