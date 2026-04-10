package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cal1604/internal/api/dto"
	"cal1604/internal/application/deviceconnect"
)

type deviceConnectConfigResponse struct {
	ConnectAttemptTimeoutMs    int `json:"connectAttemptTimeoutMs"`
	ConnectMaxAttempts         int `json:"connectMaxAttempts"`
	ConnectInitialBackoffMs    int `json:"connectInitialBackoffMs"`
	ConnectMaxBackoffMs        int `json:"connectMaxBackoffMs"`
	DisconnectAttemptTimeoutMs int `json:"disconnectAttemptTimeoutMs"`
	DisconnectMaxAttempts      int `json:"disconnectMaxAttempts"`
	DisconnectInitialBackoffMs int `json:"disconnectInitialBackoffMs"`
	DisconnectMaxBackoffMs     int `json:"disconnectMaxBackoffMs"`
}

func TestGetDeviceConnectConfig(t *testing.T) {
	router := NewRouterWithConnectConfig(nil, deviceconnect.Config{
		ConnectAttemptTimeout:    1800 * time.Millisecond,
		ConnectMaxAttempts:       5,
		ConnectInitialBackoff:    90 * time.Millisecond,
		ConnectMaxBackoff:        500 * time.Millisecond,
		DisconnectAttemptTimeout: 1100 * time.Millisecond,
		DisconnectMaxAttempts:    4,
		DisconnectInitialBackoff: 50 * time.Millisecond,
		DisconnectMaxBackoff:     250 * time.Millisecond,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/config/device-connect", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp dto.Response[deviceConnectConfigResponse]
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if !resp.Success {
		t.Fatalf("expected success response, got %+v", resp)
	}

	if resp.Data.ConnectAttemptTimeoutMs != 1800 || resp.Data.ConnectMaxAttempts != 5 {
		t.Fatalf("unexpected connect config payload: %+v", resp.Data)
	}

	if resp.Data.DisconnectAttemptTimeoutMs != 1100 || resp.Data.DisconnectMaxAttempts != 4 {
		t.Fatalf("unexpected disconnect config payload: %+v", resp.Data)
	}
}
