package http

import (
	"encoding/json"
	"net/http"

	"cal1604/internal/domain"
	apperrors "cal1604/internal/errors"
)

type measurementStartRequest struct {
	Channels []int `json:"channels"`
}

func (s *apiServer) measurementGeneratePointsHandler(w http.ResponseWriter, _ *http.Request) {
	points, err := s.measurementService.GeneratePressurePoints()
	if err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, points)
}

func (s *apiServer) measurementPointsHandler(w http.ResponseWriter, _ *http.Request) {
	writeSuccess(w, http.StatusOK, s.measurementService.GetPoints())
}

func (s *apiServer) measurementStateHandler(w http.ResponseWriter, _ *http.Request) {
	writeSuccess(w, http.StatusOK, map[string]string{"state": string(s.measurementService.State())})
}

func (s *apiServer) measurementStartHandler(w http.ResponseWriter, r *http.Request) {
	req, err := decodeJSON[measurementStartRequest](r)
	if err != nil {
		writeError(w, err)
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
		_ = s.measurementService.Stop()
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"state": string(s.measurementService.State())})
}

func (s *apiServer) measurementPauseHandler(w http.ResponseWriter, _ *http.Request) {
	if err := s.measurementService.Pause(); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"state": "paused"})
}

func (s *apiServer) measurementStopHandler(w http.ResponseWriter, _ *http.Request) {
	if err := s.measurementService.Stop(); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"state": "idle"})
}

func (s *apiServer) measurementDataHandler(w http.ResponseWriter, _ *http.Request) {
	rows, total := s.measurementService.GetData()
	writeSuccess(w, http.StatusOK, map[string]any{"rows": rows, "total": total})
}

func (s *apiServer) measurementExportHandler(w http.ResponseWriter, r *http.Request) {
	var req exportReportRequest
	if r.Method == http.MethodPost {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	if req.OutputPath == "" {
		req.OutputPath = r.URL.Query().Get("outputPath")
	}

	if req.OutputPath == "" {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	points := s.measurementService.GetPoints()
	config := s.measurementService.GetConfig()

	if err := s.reportService.ExportMeasurementReport(r.Context(), points, config, req.OutputPath); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"status": "ok", "path": req.OutputPath})
}

func (s *apiServer) measurementGetAlarmConfigHandler(w http.ResponseWriter, _ *http.Request) {
	cfg := s.measurementService.GetAlarmConfig()
	writeSuccess(w, http.StatusOK, cfg)
}

func (s *apiServer) measurementSetAlarmConfigHandler(w http.ResponseWriter, r *http.Request) {
	cfg, err := decodeJSON[domain.AlarmConfig](r)
	if err != nil {
		writeError(w, err)
		return
	}
	s.measurementService.SetAlarmConfig(cfg)
	writeSuccess(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (s *apiServer) measurementAlarmResolveHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Decision string `json:"decision"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	if err := s.measurementService.ResolveAlarm(req.Decision); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"status": "resolved"})
}

func (s *apiServer) measurementAlarmPendingHandler(w http.ResponseWriter, _ *http.Request) {
	writeSuccess(w, http.StatusOK, map[string]bool{"pending": s.measurementService.IsAlarmPending()})
}

func (s *apiServer) measurementAutoCollectHandler(w http.ResponseWriter, r *http.Request) {
	if err := s.measurementService.RunAutoCollection(r.Context()); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"state": string(s.measurementService.State())})
}

// measurementManualStartHandler 仅启动工作流（进入 ready 状态），不启动实时采样。
// 手动模式使用此端点，允许后续手动打压或直接采集。
func (s *apiServer) measurementManualStartHandler(w http.ResponseWriter, r *http.Request) {
	req, err := decodeJSON[measurementStartRequest](r)
	if err != nil {
		writeError(w, err)
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

	writeSuccess(w, http.StatusOK, map[string]string{"state": string(s.measurementService.State())})
}

func (s *apiServer) measurementManualPressurizeHandler(w http.ResponseWriter, r *http.Request) {
	req, err := decodeJSON[struct {
		PointIndex int `json:"pointIndex"`
	}](r)
	if err != nil {
		writeError(w, err)
		return
	}

	if err := s.measurementService.ManualPressurize(r.Context(), req.PointIndex); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"state": string(s.measurementService.State())})
}

func (s *apiServer) measurementManualCollectHandler(w http.ResponseWriter, r *http.Request) {
	req, err := decodeJSON[struct {
		PointIndex int `json:"pointIndex"`
	}](r)
	if err != nil {
		writeError(w, err)
		return
	}

	if err := s.measurementService.ManualCollect(r.Context(), req.PointIndex); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"state": string(s.measurementService.State())})
}
