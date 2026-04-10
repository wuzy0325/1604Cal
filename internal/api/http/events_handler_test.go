package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSSEEndpointStreamsEvents(t *testing.T) {
	router := NewRouter()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/events/stream?sessionId=s1", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		router.ServeHTTP(rec, req)
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)
	publishEvent("session.state.changed", map[string]any{"state": "ready"})

	time.Sleep(40 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("sse handler did not exit after context cancellation")
	}

	if !strings.Contains(rec.Body.String(), "session.state.changed") {
		t.Fatalf("expected sse stream body to contain event type, got %q", rec.Body.String())
	}
}
