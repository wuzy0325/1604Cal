package http

import (
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
	req, err := decodeJSON[setDevicesRequest](r)
	if err != nil {
		writeError(w, err)
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
	req, err := decodeJSON[setMeasureDeviceRequest](r)
	if err != nil {
		writeError(w, err)
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
	pressure, err := s.sessionService.ReadPressure(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, pressureResponse{Pressure: pressure})
}

func (s *apiServer) sessionReadStabilityHandler(w http.ResponseWriter, r *http.Request) {
	stable, err := s.sessionService.ReadStability(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, stabilityResponse{Stable: stable})
}

func (s *apiServer) sessionReadMeasureDataHandler(w http.ResponseWriter, r *http.Request) {
	data, err := s.sessionService.ReadMeasureData(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]any{"data": data})
}

func (s *apiServer) sessionGetValveHandler(w http.ResponseWriter, r *http.Request) {
	status, err := s.sessionService.ReadValveStatus(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, valveResponse{Status: status})
}

func (s *apiServer) sessionSetValveHandler(w http.ResponseWriter, r *http.Request) {
	req, err := decodeJSON[setValveRequest](r)
	if err != nil {
		writeError(w, err)
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
}

func (s *apiServer) sessionCalibrateZeroHandler(w http.ResponseWriter, r *http.Request) {
	req, err := decodeJSON[calibrateZeroRequest](r)
	if err != nil {
		writeError(w, err)
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
	req, err := decodeJSON[calibrateFullScaleRequest](r)
	if err != nil {
		writeError(w, err)
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

func (s *apiServer) sessionGetMeasureUnitHandler(w http.ResponseWriter, r *http.Request) {
	unit, err := s.sessionService.ReadMeasureUnit(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, measureUnitResponse{Unit: unit})
}

func (s *apiServer) sessionSetMeasureUnitHandler(w http.ResponseWriter, r *http.Request) {
	req, err := decodeJSON[setMeasureUnitRequest](r)
	if err != nil {
		writeError(w, err)
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
}

func (s *apiServer) sessionReadDeviceInfoHandler(w http.ResponseWriter, r *http.Request) {
	info, err := s.sessionService.ReadDeviceInfo(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]any{"info": info})
}

func (s *apiServer) sessionResetDeviceHandler(w http.ResponseWriter, r *http.Request) {
	if err := s.sessionService.ResetDevice(r.Context()); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"status": "ok"})
}
