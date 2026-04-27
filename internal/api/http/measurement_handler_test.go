package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cal1604/internal/api/dto"
	"cal1604/internal/application/deviceconnect"
	"cal1604/internal/config"
)

type measurementStateResponse struct {
	State string `json:"state"`
}

type measurementDataResponse struct {
	Rows  []measurementRow `json:"rows"`
	Total int              `json:"total"`
}

type measurementRow struct {
	Timestamp string             `json:"timestamp"`
	Channels  map[string]float64 `json:"channels"`
}

type measurementPointResponse struct {
	ID             string  `json:"id"`
	Index          int     `json:"index"`
	TargetPressure float64 `json:"targetPressure"`
	Direction      string  `json:"direction"`
	Status         string  `json:"status"`
}

func TestMeasurementStartCreatesWorkflowSession(t *testing.T) {
	router, _ := newSessionRouterWithMeasureDriverAndFakeDriver(t)

	configReq := httptest.NewRequest(http.MethodPost, "/api/v1/config/measurement", bytes.NewReader([]byte(`{
		"minPressure": 0,
		"maxPressure": 20,
		"pointCount": 3,
		"precision": 2,
		"averageCount": 1,
		"stableDurationMs": 1000,
		"precisionLevel": 0.05,
		"pressureMode": "single",
		"controlMode": "manual"
	}`)))
	configReq.Header.Set("Content-Type", "application/json")
	configRec := httptest.NewRecorder()
	router.ServeHTTP(configRec, configReq)
	if configRec.Code != http.StatusOK {
		t.Fatalf("expected config update status 200, got %d", configRec.Code)
	}

	generateReq := httptest.NewRequest(http.MethodPost, "/api/v1/measurement/points/generate", nil)
	generateRec := httptest.NewRecorder()
	router.ServeHTTP(generateRec, generateReq)
	if generateRec.Code != http.StatusOK {
		t.Fatalf("expected points generate status 200, got %d", generateRec.Code)
	}

	startState := callMeasurementStateEndpoint(t, router, http.MethodPost, "/api/v1/measurement/start", `{"channels":[1,2]}`)
	if startState != "collecting" {
		t.Fatalf("expected start state collecting, got %s", startState)
	}

	if state := callMeasurementStateEndpoint(t, router, http.MethodGet, "/api/v1/measurement/state", ""); state != "collecting" {
		t.Fatalf("expected current measurement state collecting, got %s", state)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/measurement/points", nil)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected points list status 200, got %d", listRec.Code)
	}

	var listResp dto.Response[[]measurementPointResponse]
	if err := json.NewDecoder(listRec.Body).Decode(&listResp); err != nil {
		t.Fatalf("decode measurement point list response: %v", err)
	}
	if len(listResp.Data) != 3 {
		t.Fatalf("expected 3 measurement points, got %d", len(listResp.Data))
	}

	if state := callMeasurementStateEndpoint(t, router, http.MethodPost, "/api/v1/measurement/stop", ""); state != "idle" {
		t.Fatalf("expected stop state idle, got %s", state)
	}
}

func TestMeasurementStartRejectsEmptyChannels(t *testing.T) {
	router := newSessionRouterWithMeasureDriver(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/measurement/start", bytes.NewReader([]byte(`{"channels":[]}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for empty channels, got %d", rec.Code)
	}
}

func TestMeasurementStartRequiresBoundMeasureDevice(t *testing.T) {
	router := NewRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/measurement/start", bytes.NewReader([]byte(`{"channels":[1]}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status 409 when measure device not bound, got %d", rec.Code)
	}
}

func TestMeasurementStartRequiresGeneratedPoints(t *testing.T) {
	router := newSessionRouterWithMeasureDriver(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/measurement/start", bytes.NewReader([]byte(`{"channels":[1]}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 when points not generated, got %d", rec.Code)
	}
}

func TestMeasurementGeneratePointsEndpointUsesMeasurementConfig(t *testing.T) {
	appCfg := config.Default()
	router := NewRouterWithRuntimeConfig(
		nil,
		deviceconnect.DefaultConfig(),
		CalibrationRuntimeConfig{},
		"",
		appCfg,
	)

	updatePayload := []byte(`{
		"minPressure": 0,
		"maxPressure": 100,
		"pointCount": 5,
		"precision": 2,
		"averageCount": 1,
		"stableDurationMs": 5000,
		"precisionLevel": 0.05,
		"pressureMode": "roundTrip",
		"controlMode": "auto"
	}`)

	configReq := httptest.NewRequest(http.MethodPost, "/api/v1/config/measurement", bytes.NewReader(updatePayload))
	configReq.Header.Set("Content-Type", "application/json")
	configRec := httptest.NewRecorder()
	router.ServeHTTP(configRec, configReq)

	if configRec.Code != http.StatusOK {
		t.Fatalf("expected config update status 200, got %d", configRec.Code)
	}

	generateReq := httptest.NewRequest(http.MethodPost, "/api/v1/measurement/points/generate", nil)
	generateRec := httptest.NewRecorder()
	router.ServeHTTP(generateRec, generateReq)

	if generateRec.Code != http.StatusOK {
		t.Fatalf("expected generate points status 200, got %d", generateRec.Code)
	}

	var resp dto.Response[[]measurementPointResponse]
	if err := json.NewDecoder(generateRec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode measurement points response: %v", err)
	}

	if len(resp.Data) != 9 {
		t.Fatalf("expected 9 measurement points (5 forward + 4 backward), got %d", len(resp.Data))
	}

	if resp.Data[0].Direction != "forward" || resp.Data[len(resp.Data)-1].Direction != "backward" {
		t.Fatalf("unexpected point directions: %+v", resp.Data)
	}

	// verify forward: [0, 25, 50, 75, 100], backward: [100, 75, 50, 25]
	expected := []float64{0, 25, 50, 75, 100, 100, 75, 50, 25}
	for i, exp := range expected {
		if resp.Data[i].TargetPressure != exp {
			t.Fatalf("points[%d].TargetPressure = %v, want %v", i, resp.Data[i].TargetPressure, exp)
		}
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/measurement/points", nil)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("expected list points status 200, got %d", listRec.Code)
	}

	var listResp dto.Response[[]measurementPointResponse]
	if err := json.NewDecoder(listRec.Body).Decode(&listResp); err != nil {
		t.Fatalf("decode listed measurement points response: %v", err)
	}

	if len(listResp.Data) != len(resp.Data) {
		t.Fatalf("expected listed points count %d, got %d", len(resp.Data), len(listResp.Data))
	}
}

func TestMeasurementGeneratePointsRejectsInvalidConfig(t *testing.T) {
	appCfg := config.Default()
	router := NewRouterWithRuntimeConfig(
		nil,
		deviceconnect.DefaultConfig(),
		CalibrationRuntimeConfig{},
		"",
		appCfg,
	)

	invalidPayload := []byte(`{
		"minPressure": 0,
		"maxPressure": 100,
		"pointCount": 1,
		"precision": 2,
		"averageCount": 1,
		"stableDurationMs": 5000,
		"precisionLevel": 0.05,
		"pressureMode": "single",
		"controlMode": "auto"
	}`)

	configReq := httptest.NewRequest(http.MethodPost, "/api/v1/config/measurement", bytes.NewReader(invalidPayload))
	configReq.Header.Set("Content-Type", "application/json")
	configRec := httptest.NewRecorder()
	router.ServeHTTP(configRec, configReq)

	if configRec.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid config status 400, got %d", configRec.Code)
	}
}

func waitForMeasurementData(t *testing.T, router http.Handler, minRows int, timeout time.Duration) measurementDataResponse {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for {
		resp := callMeasurementDataEndpoint(t, router)
		if resp.Total >= minRows {
			return resp
		}

		if time.Now().After(deadline) {
			t.Fatalf("timeout waiting measurement data rows >= %d, latest=%d", minRows, resp.Total)
		}

		time.Sleep(120 * time.Millisecond)
	}
}

func callMeasurementStateEndpoint(t *testing.T, router http.Handler, method, path, body string) string {
	t.Helper()

	var reqBody *bytes.Reader
	if body == "" {
		reqBody = bytes.NewReader(nil)
	} else {
		reqBody = bytes.NewReader([]byte(body))
	}

	req := httptest.NewRequest(method, path, reqBody)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("request %s %s failed with status %d", method, path, rec.Code)
	}

	var resp dto.Response[measurementStateResponse]
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode measurement state response: %v", err)
	}

	return resp.Data.State
}

func callMeasurementDataEndpoint(t *testing.T, router http.Handler) measurementDataResponse {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/measurement/data", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("request GET /api/v1/measurement/data failed with status %d", rec.Code)
	}

	var resp dto.Response[measurementDataResponse]
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode measurement data response: %v", err)
	}

	return resp.Data
}
