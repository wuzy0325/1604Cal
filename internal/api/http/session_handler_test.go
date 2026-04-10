package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cal1604/internal/api/dto"
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

func TestSessionStartPauseResumeStopFlow(t *testing.T) {
	router := NewRouter()

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

	if got := call(http.MethodPost, "/api/v1/sessions/start").State; got != "pressurizing" {
		t.Fatalf("expected start state pressurizing, got %s", got)
	}

	if got := call(http.MethodPost, "/api/v1/sessions/pause").State; got != "paused" {
		t.Fatalf("expected pause state paused, got %s", got)
	}

	if got := call(http.MethodPost, "/api/v1/sessions/resume").State; got != "pressurizing" {
		t.Fatalf("expected resume state pressurizing, got %s", got)
	}

	if got := call(http.MethodPost, "/api/v1/sessions/stop").State; got != "stopped" {
		t.Fatalf("expected stop state stopped, got %s", got)
	}
}
