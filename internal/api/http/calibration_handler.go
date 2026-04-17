package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"cal1604/internal/application/calibration"
	apperrors "cal1604/internal/errors"
)

type setDevicesRequest struct {
	MeasureDeviceID  string `json:"measureDeviceId"`
	PressureDeviceID string `json:"pressureDeviceId"`
}

type setMeasureDeviceRequest struct {
	MeasureDeviceID string `json:"measureDeviceId"`
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

type pressureResponse struct {
	Pressure float64 `json:"pressure"`
}

type stabilityResponse struct {
	Stable bool `json:"stable"`
}

type channelsResponse struct {
	Channels []int `json:"channels"`
}

type setValveRequest struct {
	Status string `json:"status"`
}

type valveResponse struct {
	Status string `json:"status"`
}

type measureUnitResponse struct {
	Unit string `json:"unit"`
}

type setMeasureUnitRequest struct {
	Unit string `json:"unit"`
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

	if req.MeasureDeviceID == "" || req.PressureDeviceID == "" {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	if err := s.calibrationService.SetDevices(req.MeasureDeviceID, req.PressureDeviceID); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *apiServer) calibrationSetMeasureDeviceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req setMeasureDeviceRequest
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

	if err := s.calibrationService.SetMeasureDevice(req.MeasureDeviceID); err != nil {
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
	writeSuccess(w, http.StatusOK, channelsResponse{Channels: req.Channels})
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

	points := s.calibrationService.GeneratePressurePoints()
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

	if err := s.calibrationService.Fit(r.Context()); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"status": "completed"})
}

func (s *apiServer) calibrationReadPressureHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	pressure, err := s.calibrationService.ReadCurrentPressure(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, pressureResponse{Pressure: pressure})
}

func (s *apiServer) calibrationReadStabilityHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	stable, err := s.calibrationService.ReadStability(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, stabilityResponse{Stable: stable})
}

func (s *apiServer) calibrationReadMeasureDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	data, err := s.calibrationService.ReadMeasureData(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]any{"data": data})
}

func (s *apiServer) calibrationValveHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		status, err := s.calibrationService.ReadValveStatus(r.Context())
		if err != nil {
			writeError(w, err)
			return
		}
		writeSuccess(w, http.StatusOK, valveResponse{Status: status})

	case http.MethodPost:
		var req setValveRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			writeError(w, apperrors.ErrInvalidArgument)
			return
		}
		if req.Status == "" {
			writeError(w, apperrors.ErrInvalidArgument)
			return
		}

		if err := s.calibrationService.SetValveStatus(r.Context(), req.Status); err != nil {
			writeError(w, err)
			return
		}
		writeSuccess(w, http.StatusOK, valveResponse{Status: req.Status})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *apiServer) calibrationReadMeasureUnitHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		unit, err := s.calibrationService.ReadMeasureUnit(r.Context())
		if err != nil {
			writeError(w, err)
			return
		}
		writeSuccess(w, http.StatusOK, measureUnitResponse{Unit: unit})

	case http.MethodPost:
		var req setMeasureUnitRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			writeError(w, apperrors.ErrInvalidArgument)
			return
		}
		if req.Unit == "" {
			writeError(w, apperrors.ErrInvalidArgument)
			return
		}

		if err := s.calibrationService.SetMeasureUnit(r.Context(), req.Unit); err != nil {
			writeError(w, err)
			return
		}
		writeSuccess(w, http.StatusOK, measureUnitResponse{Unit: req.Unit})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *apiServer) calibrationReadDeviceInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	info, err := s.calibrationService.ReadDeviceInfo(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]any{"info": info})
}

func (s *apiServer) calibrationResetDeviceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := s.calibrationService.ResetMeasureDevice(r.Context()); err != nil {
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
