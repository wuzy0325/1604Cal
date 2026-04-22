package http

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"

	"cal1604/internal/application/calibration"
	"cal1604/internal/application/deviceconnect"
	"cal1604/internal/config"
	"cal1604/internal/application/measurement"
	"cal1604/internal/application/multipress"
	"cal1604/internal/application/session"
	"cal1604/internal/domain"
	apperrors "cal1604/internal/errors"
	"cal1604/internal/report"
	"cal1604/internal/workflow"
)

// deviceManager 定义 apiServer 对设备管理器的依赖接口。
// 同时兼容内存版 DeviceManager 和持久化版 PersistentDeviceManager。
type deviceManager interface {
	Upsert(dev domain.Device)
	UpdateStatus(id string, status domain.DeviceStatus) bool
	Delete(id string)
	Get(id string) (domain.Device, bool)
	List() []domain.Device
	CheckUnitConsistency() (bool, []string)
}

type apiServer struct {
	deviceManager      deviceManager
	sessionMachine     *workflow.SessionMachine
	deviceConnector    deviceConnector
	connectConfig      deviceconnect.Config
	calibrationService *calibration.Service
	multipressService  *multipress.Service
	sessionService     *session.Service
	measurementService *measurement.Service
	reportService     *report.Service
	configPath         string
	appConfig          *config.AppConfig
}

type deviceConnector interface {
	Connect(ctx context.Context, id string) (domain.Device, error)
	Disconnect(ctx context.Context, id string) (domain.Device, error)
}

type upsertDeviceRequest struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Model  string `json:"model"`
	Host   string `json:"host"`
	Port   int    `json:"port"`
	Unit   string `json:"unit"`
	Status string `json:"status"`
}

type setDeviceStatusRequest struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type unitConsistencyPayload struct {
	Consistent bool     `json:"consistent"`
	Conflicts  []string `json:"conflicts"`
}

func (s *apiServer) devicesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		devices := s.deviceManager.List()
		if devices == nil {
			devices = make([]domain.Device, 0)
		}
		writeSuccess(w, http.StatusOK, devices)
	case http.MethodPost:
		s.handleUpsertDevice(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *apiServer) deviceStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req setDeviceStatusRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	id := strings.TrimSpace(req.ID)
	status := domain.DeviceStatus(strings.TrimSpace(req.Status))
	if id == "" || !isValidDeviceStatus(status) {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	if ok := s.deviceManager.UpdateStatus(id, status); !ok {
		writeError(w, apperrors.ErrNotFound)
		return
	}

	dev, exists := s.deviceManager.Get(id)
	if !exists {
		writeError(w, apperrors.ErrNotFound)
		return
	}

	s.publishDeviceStatusChanged(dev)

	writeSuccess(w, http.StatusOK, map[string]string{"id": id, "status": string(status)})
}

func (s *apiServer) deviceConnectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id, ok := decodeDeviceIDRequest(r, w)
	if !ok {
		return
	}

	if s.deviceConnector == nil {
		writeError(w, errors.New("device connector is nil"))
		return
	}

	updated, err := s.deviceConnector.Connect(r.Context(), id)
	if errors.Is(err, apperrors.ErrNotFound) {
		writeError(w, apperrors.ErrNotFound)
		return
	}

	// 连接失败时仍返回设备快照，前端可直接展示 error 状态与失败原因。
	if err != nil {
		writeSuccess(w, http.StatusOK, updated)
		return
	}

	writeSuccess(w, http.StatusOK, updated)
}

func (s *apiServer) deviceDisconnectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id, ok := decodeDeviceIDRequest(r, w)
	if !ok {
		return
	}

	if s.deviceConnector == nil {
		writeError(w, errors.New("device connector is nil"))
		return
	}

	updated, err := s.deviceConnector.Disconnect(r.Context(), id)
	if errors.Is(err, apperrors.ErrNotFound) {
		writeError(w, apperrors.ErrNotFound)
		return
	}

	// 断连失败同样返回设备快照，便于前端统一展示错误信息。
	if err != nil {
		writeSuccess(w, http.StatusOK, updated)
		return
	}

	writeSuccess(w, http.StatusOK, updated)
}

func (s *apiServer) handleUpsertDevice(w http.ResponseWriter, r *http.Request) {
	var req upsertDeviceRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	deviceType := domain.DeviceType(strings.TrimSpace(req.Type))
	id := strings.TrimSpace(req.ID)
	host := strings.TrimSpace(req.Host)
	unit := strings.TrimSpace(req.Unit)

	if id == "" || !isValidDeviceType(deviceType) {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	if net.ParseIP(host) == nil || req.Port < 1 || req.Port > 65535 || unit == "" {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	requestedStatus := domain.DeviceStatus(strings.TrimSpace(req.Status))
	if req.Status != "" && !isValidDeviceStatus(requestedStatus) {
		writeError(w, apperrors.ErrInvalidArgument)
		return
	}

	status := requestedStatus
	old, existed := s.deviceManager.Get(id)
	if status == "" {
		if existed && old.Status != "" {
			status = old.Status
		} else {
			status = domain.DeviceStatusDisconnected
		}
	}

	dev := domain.Device{
		ID:     id,
		Name:   strings.TrimSpace(req.Name),
		Type:   deviceType,
		Model:  strings.TrimSpace(req.Model),
		Host:   host,
		Port:   req.Port,
		Unit:   unit,
		Status: status,
	}
	if existed {
		dev.LastErrorReason = old.LastErrorReason
		dev.LastErrorAt = old.LastErrorAt
	}

	s.deviceManager.Upsert(dev)
	s.publishDeviceStatusChanged(dev)
	writeSuccess(w, http.StatusOK, dev)
}

func (s *apiServer) publishDeviceStatusChanged(dev domain.Device) {
	payload := map[string]any{
		"id":     dev.ID,
		"type":   string(dev.Type),
		"status": string(dev.Status),
	}
	if dev.LastErrorReason != "" {
		payload["errorReason"] = dev.LastErrorReason
	}
	if dev.LastErrorAt != nil {
		payload["lastErrorAt"] = dev.LastErrorAt
	}

	publishEvent("device.status.changed", payload)
}

func (s *apiServer) unitConsistencyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	consistent, conflicts := s.deviceManager.CheckUnitConsistency()
	writeSuccess(w, http.StatusOK, unitConsistencyPayload{
		Consistent: consistent,
		Conflicts:  conflicts,
	})
}

func isValidDeviceType(deviceType domain.DeviceType) bool {
	return deviceType == domain.DeviceTypeMeasure || deviceType == domain.DeviceTypePressure
}

func isValidDeviceStatus(status domain.DeviceStatus) bool {
	return status == domain.DeviceStatusDisconnected ||
		status == domain.DeviceStatusConnecting ||
		status == domain.DeviceStatusConnected ||
		status == domain.DeviceStatusError
}

func decodeDeviceIDRequest(r *http.Request, w http.ResponseWriter) (string, bool) {
	var req setDeviceStatusRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, apperrors.ErrInvalidArgument)
		return "", false
	}

	id := strings.TrimSpace(req.ID)
	if id == "" {
		writeError(w, apperrors.ErrInvalidArgument)
		return "", false
	}

	return id, true
}
