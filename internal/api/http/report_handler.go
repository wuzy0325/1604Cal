package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	apperrors "cal1604/internal/errors"
	"cal1604/internal/report"
)

type reportTemplateSelection struct {
	Filename string `json:"filename"`
}

func (s *apiServer) reportTemplateSelectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	pointsText := strings.TrimSpace(r.URL.Query().Get("points"))
	mode := strings.TrimSpace(r.URL.Query().Get("mode"))
	if pointsText == "" || mode == "" {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	points, err := strconv.Atoi(pointsText)
	if err != nil {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	filename, err := report.SelectTemplate(points, mode)
	if err != nil {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	writeSuccess(w, http.StatusOK, reportTemplateSelection{Filename: filename})
}

type exportReportRequest struct {
	OutputPath string `json:"outputPath"`
}

// exportReportHandler 根据当前校准会话导出校准报告。
func (s *apiServer) exportReportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req exportReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}
	if req.OutputPath == "" {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	session := s.calibrationService.GetCalibrationSession()
	if session == nil {
		writeError(w, fmt.Errorf("no active calibration session"))
		return
	}

	if err := s.reportService.ExportReport(r.Context(), session, req.OutputPath); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"status": "ok", "path": req.OutputPath})
}

// listTemplatesHandler 返回可用的报告模板列表。
func (s *apiServer) listTemplatesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	templates, err := s.reportService.GetTemplates()
	if err != nil {
		writeError(w, err)
		return
	}

	if templates == nil {
		templates = []string{}
	}
	writeSuccess(w, http.StatusOK, map[string]any{"templates": templates})
}
