package http

import (
	"encoding/json"
	"net/http"
	"strings"

	apperrors "cal1604/internal/errors"
)

type setMeasureDeviceRequest struct {
	MeasureDeviceID string `json:"measureDeviceId"`
}

type pressureResponse struct {
	Pressure float64 `json:"pressure"`
}

type stabilityResponse struct {
	Stable bool `json:"stable"`
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

type calibrateZeroRequest struct {
	Channels []int `json:"channels"`
}

type calibrateFullScaleRequest struct {
	Channels       []int   `json:"channels"`
	FullScaleValue float64 `json:"fullScaleValue"`
}

func (s *apiServer) sessionSetDevicesHandler(w http.ResponseWriter, r *http.Request) {
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

	if err := s.sessionService.BindDevices(req.MeasureDeviceID, req.PressureDeviceID); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *apiServer) sessionSetMeasureDeviceHandler(w http.ResponseWriter, r *http.Request) {
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

	if err := s.sessionService.BindMeasureDevice(req.MeasureDeviceID); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *apiServer) sessionReadPressureHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	pressure, err := s.sessionService.ReadPressure(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, pressureResponse{Pressure: pressure})
}

func (s *apiServer) sessionReadStabilityHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	stable, err := s.sessionService.ReadStability(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, stabilityResponse{Stable: stable})
}

func (s *apiServer) sessionReadMeasureDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	data, err := s.sessionService.ReadMeasureData(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]any{"data": data})
}

func (s *apiServer) sessionValveHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		status, err := s.sessionService.ReadValveStatus(r.Context())
		if err != nil {
			writeError(w, err)
			return
		}
		writeSuccess(w, http.StatusOK, valveResponse{Status: status})

	case http.MethodPut, http.MethodPost:
		var req setValveRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			writeError(w, apperrors.ErrInvalidArgument)
			return
		}

		normalizedStatus, ok := normalizeValveStatus(req.Status)
		if !ok {
			writeError(w, apperrors.ErrInvalidArgument)
			return
		}

		if err := s.sessionService.SetValveStatus(r.Context(), normalizedStatus); err != nil {
			writeError(w, err)
			return
		}
		writeSuccess(w, http.StatusOK, valveResponse{Status: normalizedStatus})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *apiServer) sessionCalibrateZeroHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req calibrateZeroRequest
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

	values, err := s.sessionService.CalibrateZero(r.Context(), req.Channels)
	if err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]any{"values": values})
}

func (s *apiServer) sessionCalibrateFullScaleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req calibrateFullScaleRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}
	if len(req.Channels) == 0 || req.FullScaleValue <= 0 {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	values, err := s.sessionService.CalibrateFullScale(r.Context(), req.Channels, req.FullScaleValue)
	if err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]any{"values": values})
}

func normalizeValveStatus(status string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "calibration":
		return "calibration", true
	case "measurement":
		return "measurement", true
	default:
		return "", false
	}
}

func (s *apiServer) sessionMeasureUnitHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		unit, err := s.sessionService.ReadMeasureUnit(r.Context())
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

		if err := s.sessionService.SetMeasureUnit(r.Context(), req.Unit); err != nil {
			writeError(w, err)
			return
		}
		writeSuccess(w, http.StatusOK, measureUnitResponse(req))

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *apiServer) sessionReadDeviceInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	info, err := s.sessionService.ReadDeviceInfo(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]any{"info": info})
}

func (s *apiServer) sessionResetDeviceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := s.sessionService.ResetDevice(r.Context()); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"status": "ok"})
}
