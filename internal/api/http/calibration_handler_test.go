package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cal1604/internal/api/dto"
)

type calibrationPointResponse struct {
	Index          int     `json:"index"`
	TargetPressure float64 `json:"targetPressure"`
	Status         string  `json:"status"`
}

func TestCalibrationConfigAcceptsControlAndPressureMode(t *testing.T) {
	router := NewRouter()

	payload := map[string]any{
		"channels":       []int{1, 2},
		"pressurePoints": 5,
		"averageCount":   3,
		"minPressure":    0,
		"maxPressure":    100,
		"stableWaitMs":   3000,
		"controlMode":    "auto",
		"pressureMode":   "return",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/calibration/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp dto.Response[map[string]string]
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if !resp.Success {
		t.Fatalf("expected success response, got %+v", resp)
	}

	generateReq := httptest.NewRequest(http.MethodPost, "/api/v1/calibration/points/generate", nil)
	generateRec := httptest.NewRecorder()
	router.ServeHTTP(generateRec, generateReq)

	if generateRec.Code != http.StatusOK {
		t.Fatalf("expected generate status 200, got %d", generateRec.Code)
	}

	var generateResp dto.Response[[]calibrationPointResponse]
	if err := json.NewDecoder(generateRec.Body).Decode(&generateResp); err != nil {
		t.Fatalf("decode generate response: %v", err)
	}

	if !generateResp.Success {
		t.Fatalf("expected generate success response, got %+v", generateResp)
	}

	if len(generateResp.Data) != 5 {
		t.Fatalf("expected 5 generated points, got %d", len(generateResp.Data))
	}
}
