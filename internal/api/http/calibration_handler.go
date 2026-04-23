package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"cal1604/internal/application/calibration"
	"cal1604/internal/domain"
	apperrors "cal1604/internal/errors"
)

type setDevicesRequest struct {
	MeasureDeviceID  string `json:"measureDeviceId"`
	PressureDeviceID string `json:"pressureDeviceId"`
}

type setConfigRequest struct {
	Channels       []int   `json:"channels"`
	PressurePoints int     `json:"pressurePoints"`
	AverageCount   int     `json:"averageCount"`
	MinPressure    float64 `json:"minPressure"`
	MaxPressure    float64 `json:"maxPressure"`
	StableWaitMs   int     `json:"stableWaitMs"`
	ControlMode    string  `json:"controlMode,omitempty"`
	PressureMode   string  `json:"pressureMode,omitempty"`
}

type setChannelsRequest struct {
	Channels []int `json:"channels"`
}

type pointIndexRequest struct {
	PointIndex int `json:"pointIndex"`
}

type channelsResponse struct {
	Channels []int `json:"channels"`
}

type alarmDecisionRequest struct {
	Decision string `json:"decision"`
}

func (s *apiServer) calibrationSetDevicesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req setDevicesRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	if req.MeasureDeviceID == "" {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	if err := s.calibrationService.SetDevices(req.MeasureDeviceID, req.PressureDeviceID); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *apiServer) calibrationSetConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req setConfigRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	config := calibration.CalibrationConfig{
		Channels:       req.Channels,
		PressurePoints: req.PressurePoints,
		AverageCount:   req.AverageCount,
		MinPressure:    req.MinPressure,
		MaxPressure:    req.MaxPressure,
		StableWaitMs:   req.StableWaitMs,
		ControlMode:    req.ControlMode,
		PressureMode:   req.PressureMode,
	}
	if config.PressurePoints < 2 || config.PressurePoints > 6 {
		writeError(w, fmt.Errorf("pressure points must be between 2 and 6, got %d", config.PressurePoints))
		return
	}
	s.calibrationService.SetConfig(config)

	writeSuccess(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *apiServer) calibrationSetChannelsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req setChannelsRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	s.calibrationService.SetChannels(req.Channels)
	writeSuccess(w, http.StatusOK, channelsResponse(req))
}

func (s *apiServer) calibrationGetChannelsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	channels := s.calibrationService.GetChannels()
	writeSuccess(w, http.StatusOK, channelsResponse{Channels: channels})
}

func (s *apiServer) calibrationGeneratePointsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	points, err := s.calibrationService.GeneratePressurePoints()
	if err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, points)
}

func (s *apiServer) calibrationGetPointsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	points := s.calibrationService.GetPressurePoints()
	writeSuccess(w, http.StatusOK, points)
}

func (s *apiServer) calibrationPressurizeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	pointIndex, ok := decodePointIndexRequest(r, w)
	if !ok {
		return
	}

	if err := s.calibrationService.Pressurize(r.Context(), pointIndex); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"status": "pressurizing"})
}

func (s *apiServer) calibrationCollectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	pointIndex, ok := decodePointIndexRequest(r, w)
	if !ok {
		return
	}

	data, err := s.calibrationService.Collect(r.Context(), pointIndex)
	if err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]any{"data": data})
}

func (s *apiServer) calibrationFitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	result, err := s.calibrationService.Fit(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, result)
}

func (s *apiServer) calibrationResolveAlarmHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req alarmDecisionRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	if err := s.calibrationService.ResolveAlarm(req.Decision); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *apiServer) calibrationRetryPointHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	pointIndex, ok := decodePointIndexRequest(r, w)
	if !ok {
		return
	}

	if err := s.calibrationService.RetryPoint(r.Context(), pointIndex); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"status": "ok"})
}

func decodePointIndexRequest(r *http.Request, w http.ResponseWriter) (int, bool) {
	var req pointIndexRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		// 尝试从查询参数读取
		idxStr := r.URL.Query().Get("pointIndex")
		if idxStr != "" {
			idx, err := strconv.Atoi(idxStr)
			if err != nil || idx < 1 {
				writeError(w, apperrors.ErrInvalidArgument)
				return 0, false
			}
			return idx, true
		}
		writeError(w, apperrors.ErrInvalidArgument)
		return 0, false
	}

	if req.PointIndex < 1 {
		writeError(w, apperrors.ErrInvalidArgument)
		return 0, false
	}

	return req.PointIndex, true
}

// calibrationGetAlarmConfigHandler 返回当前报警配置。
func (s *apiServer) calibrationGetAlarmConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	config := s.calibrationService.GetAlarmConfig()
	writeSuccess(w, http.StatusOK, config)
}

// calibrationSetAlarmConfigHandler 更新报警配置。
func (s *apiServer) calibrationSetAlarmConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var config domain.AlarmConfig
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&config); err != nil {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	s.calibrationService.SetAlarmConfig(config)
	writeSuccess(w, http.StatusOK, map[string]string{"status": "ok"})
}

// calibrationGetSessionHandler 返回当前校准会话。
func (s *apiServer) calibrationGetSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	session := s.calibrationService.GetCalibrationSession()
	writeSuccess(w, http.StatusOK, session)
}

type manualPressurizeRequest struct {
	TargetPressure float64 `json:"targetPressure"`
}

func (s *apiServer) calibrationManualPressurizeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req manualPressurizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	if err := s.calibrationService.ManualPressurize(r.Context(), req.TargetPressure); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *apiServer) calibrationManualCollectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	data, err := s.calibrationService.ManualCollect(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]any{"data": data})
}
