package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cal1604/internal/api/dto"
	"cal1604/internal/application/deviceconnect"
	"cal1604/internal/device"
	"cal1604/internal/device/manager"
	"cal1604/internal/domain"
)

type sessionStateResponse struct {
	State string `json:"state"`
}

func TestSessionInitialStateIsIdle(t *testing.T) {
	router := NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/state", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp dto.Response[sessionStateResponse]
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Data.State != "idle" {
		t.Fatalf("expected state idle, got %s", resp.Data.State)
	}
}

func TestSessionStartStopRestartFlow(t *testing.T) {
	router := newSessionRouterWithMeasureDriver(t)

	call := func(method, path string) sessionStateResponse {
		t.Helper()
		req := httptest.NewRequest(method, path, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("request %s %s failed with %d", method, path, rec.Code)
		}

		var resp dto.Response[sessionStateResponse]
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode response: %v", err)
		}

		return resp.Data
	}

	if got := call(http.MethodPost, "/api/v1/sessions/start").State; got != "ready" {
		t.Fatalf("expected start state ready, got %s", got)
	}

	if got := call(http.MethodPost, "/api/v1/sessions/stop").State; got != "stopped" {
		t.Fatalf("expected stop state stopped, got %s", got)
	}

	if got := call(http.MethodPost, "/api/v1/sessions/start").State; got != "ready" {
		t.Fatalf("expected restart state ready, got %s", got)
	}
}

func TestSessionStartWithoutDeviceRejected(t *testing.T) {
	router := NewRouter()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/start", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", rec.Code)
	}
}

func newSessionRouterWithMeasureDriver(t *testing.T) http.Handler {
	t.Helper()

	store := manager.NewDeviceManager()
	connector := &sessionTestConnector{
		activeDrivers: map[string]device.ConnectionDriver{
			"m1": &sessionFakeMeasureDriver{
				valveStatus: "measurement",
				unit:        "kPa",
			},
		},
	}
	router := NewRouterWithDependencies(store, connector)

	body := bytes.NewReader([]byte(`{"measureDeviceId":"m1","pressureDeviceId":""}`))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/session/devices", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("bind measure device failed, status=%d", rec.Code)
	}

	configBody := bytes.NewReader([]byte(`{
		"channels":[1],
		"pressurePoints":2,
		"averageCount":1,
		"minPressure":0,
		"maxPressure":10,
		"stableWaitMs":1000,
		"controlMode":"manual"
	}`))
	configReq := httptest.NewRequest(http.MethodPost, "/api/v1/calibration/config", configBody)
	configReq.Header.Set("Content-Type", "application/json")
	configRec := httptest.NewRecorder()
	router.ServeHTTP(configRec, configReq)
	if configRec.Code != http.StatusOK {
		t.Fatalf("set calibration config failed, status=%d", configRec.Code)
	}

	return router
}

func newSessionRouterWithMeasureDriverAndRuntimeConfig(t *testing.T, runtimeCfg CalibrationRuntimeConfig) http.Handler {
	t.Helper()

	store := manager.NewDeviceManager()
	connector := &sessionTestConnector{
		activeDrivers: map[string]device.ConnectionDriver{
			"m1": &sessionFakeMeasureDriver{
				valveStatus: "measurement",
				unit:        "kPa",
			},
		},
	}
	router := newRouter(store, connector, deviceconnect.DefaultConfig(), runtimeCfg, nil)

	bindBody := bytes.NewReader([]byte(`{"measureDeviceId":"m1","pressureDeviceId":""}`))
	bindReq := httptest.NewRequest(http.MethodPost, "/api/v1/session/devices", bindBody)
	bindReq.Header.Set("Content-Type", "application/json")
	bindRec := httptest.NewRecorder()
	router.ServeHTTP(bindRec, bindReq)
	if bindRec.Code != http.StatusOK {
		t.Fatalf("bind measure device failed, status=%d", bindRec.Code)
	}

	configBody := bytes.NewReader([]byte(`{
		"channels":[1],
		"pressurePoints":2,
		"averageCount":1,
		"minPressure":0,
		"maxPressure":10,
		"stableWaitMs":1000,
		"controlMode":"manual"
	}`))
	configReq := httptest.NewRequest(http.MethodPost, "/api/v1/calibration/config", configBody)
	configReq.Header.Set("Content-Type", "application/json")
	configRec := httptest.NewRecorder()
	router.ServeHTTP(configRec, configReq)
	if configRec.Code != http.StatusOK {
		t.Fatalf("set calibration config failed, status=%d", configRec.Code)
	}

	return router
}

func TestSessionStartRejectsWhenValveGateEnabled(t *testing.T) {
	router := newSessionRouterWithMeasureDriverAndRuntimeConfig(t, CalibrationRuntimeConfig{
		EnforceValveCalibrationGate: true,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/start", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status 409 when valve gate enabled, got %d", rec.Code)
	}
}

func TestSessionStartAllowsWhenValveGateDisabled(t *testing.T) {
	router := newSessionRouterWithMeasureDriverAndRuntimeConfig(t, CalibrationRuntimeConfig{
		EnforceValveCalibrationGate: false,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/start", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 when valve gate disabled, got %d", rec.Code)
	}
}

// sessionTestConnector 提供测试用的活动驱动，用于模拟已连接设备。
type sessionTestConnector struct {
	activeDrivers map[string]device.ConnectionDriver
}

func (c *sessionTestConnector) Connect(_ context.Context, id string) (domain.Device, error) {
	return domain.Device{ID: id, Status: domain.DeviceStatusConnected}, nil
}

func (c *sessionTestConnector) Disconnect(_ context.Context, id string) (domain.Device, error) {
	return domain.Device{ID: id, Status: domain.DeviceStatusDisconnected}, nil
}

func (c *sessionTestConnector) GetActiveDriver(id string) device.ConnectionDriver {
	if c.activeDrivers == nil {
		return nil
	}
	return c.activeDrivers[id]
}

// sessionFakeMeasureDriver 仅实现会话启动/停止测试所需的计量驱动能力。
type sessionFakeMeasureDriver struct {
	valveStatus string
	unit        string
}

func (d *sessionFakeMeasureDriver) Connect(_ context.Context) error {
	return nil
}

func (d *sessionFakeMeasureDriver) Disconnect(_ context.Context) error {
	return nil
}

func (d *sessionFakeMeasureDriver) ReadValveStatus(_ context.Context) (string, error) {
	if d.valveStatus == "" {
		return "measurement", nil
	}
	return d.valveStatus, nil
}

func (d *sessionFakeMeasureDriver) SetValveStatus(_ context.Context, status string) error {
	d.valveStatus = status
	return nil
}

func (d *sessionFakeMeasureDriver) ReadUnit(_ context.Context) (string, error) {
	if d.unit == "" {
		return "kPa", nil
	}
	return d.unit, nil
}

func (d *sessionFakeMeasureDriver) SetUnit(_ context.Context, unit string) error {
	d.unit = unit
	return nil
}

func (d *sessionFakeMeasureDriver) CollectData(_ context.Context, channels []int) ([]float64, error) {
	result := make([]float64, len(channels))
	for i := range channels {
		result[i] = 0
	}
	return result, nil
}

func (d *sessionFakeMeasureDriver) ReadDeviceInfo(_ context.Context) (map[string]string, error) {
	return map[string]string{
		"model": "fake-measure",
	}, nil
}

func (d *sessionFakeMeasureDriver) Reset(_ context.Context) error {
	return nil
}
