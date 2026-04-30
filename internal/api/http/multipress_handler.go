package http

import (
	"encoding/json"
	"log"
	"net/http"

	apperrors "cal1604/internal/errors"
)

// multipressDeviceIDRequest 多设备打压模块通用请求（仅含设备ID）。
type multipressDeviceIDRequest struct {
	DeviceID string `json:"deviceId"`
}

// multipressSetPressureRequest 设置目标压力请求。
type multipressSetPressureRequest struct {
	DeviceID        string  `json:"deviceId"`
	TargetPressure  float64 `json:"targetPressure"`
}

// multipressSetUnitRequest 设置单位请求。
type multipressSetUnitRequest struct {
	DeviceID string `json:"deviceId"`
	Unit     string `json:"unit"`
}

func (s *apiServer) multipressRegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req multipressDeviceIDRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	if req.DeviceID == "" {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	if err := s.multipressService.RegisterDevice(r.Context(), req.DeviceID); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *apiServer) multipressUnregisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req multipressDeviceIDRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	if req.DeviceID == "" {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	if err := s.multipressService.UnregisterDevice(r.Context(), req.DeviceID); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *apiServer) multipressSetPressureHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req multipressSetPressureRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	if req.DeviceID == "" {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	if err := s.multipressService.SetTargetPressure(r.Context(), req.DeviceID, req.TargetPressure); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *apiServer) multipressStopHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req multipressDeviceIDRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	if req.DeviceID == "" {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	if err := s.multipressService.Stop(r.Context(), req.DeviceID); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *apiServer) multipressExhaustHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req multipressDeviceIDRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	if req.DeviceID == "" {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	if err := s.multipressService.Exhaust(r.Context(), req.DeviceID); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *apiServer) multipressReadPressureHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	deviceID := r.URL.Query().Get("deviceId")
	if deviceID == "" {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	pressure, err := s.multipressService.ReadCurrentPressure(r.Context(), deviceID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]any{
		"pressure": pressure,
		"deviceId": deviceID,
	})
}

func (s *apiServer) multipressReadStabilityHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	deviceID := r.URL.Query().Get("deviceId")
	if deviceID == "" {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	stable, err := s.multipressService.ReadStability(r.Context(), deviceID)
	if err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]any{
		"stable":   stable,
		"deviceId": deviceID,
	})
}

func (s *apiServer) multipressUnitHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		deviceID := r.URL.Query().Get("deviceId")
		if deviceID == "" {
			writeError(w, apperrors.ErrInvalidArgument)
			return
		}

		unit, err := s.multipressService.ReadUnit(r.Context(), deviceID)
		if err != nil {
			log.Printf("[API multipressUnitHandler GET] %s error: %v", deviceID, err)
			writeError(w, err)
			return
		}
		log.Printf("[API multipressUnitHandler GET] %s → %q", deviceID, unit)

		writeSuccess(w, http.StatusOK, map[string]any{
			"unit":     unit,
			"deviceId": deviceID,
		})

	case http.MethodPost:
		var req multipressSetUnitRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			writeError(w, apperrors.ErrInvalidArgument)
			return
		}

		if req.DeviceID == "" || req.Unit == "" {
			writeError(w, apperrors.ErrInvalidArgument)
			return
		}
		log.Printf("[API multipressUnitHandler POST] %s → %q", req.DeviceID, req.Unit)

		if err := s.multipressService.SetUnit(r.Context(), req.DeviceID, req.Unit); err != nil {
			log.Printf("[API multipressUnitHandler POST] %s error: %v", req.DeviceID, err)
			writeError(w, err)
			return
		}

		log.Printf("[API multipressUnitHandler POST] %s done", req.DeviceID)
		writeSuccess(w, http.StatusOK, map[string]string{"status": "ok"})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *apiServer) multipressDevicesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	states := s.multipressService.ListDeviceStates()
	writeSuccess(w, http.StatusOK, states)
}

func (s *apiServer) multipressStopAllHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := s.multipressService.StopAll(r.Context()); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{"status": "ok"})
}
