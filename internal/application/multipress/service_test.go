package multipress

import (
	"testing"
	"time"
)

func TestStopPollingReturnsAfterStart(t *testing.T) {
	svc := NewService(nil, nil, nil)
	svc.StartPolling()

	done := make(chan struct{})
	go func() {
		svc.StopPolling()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("StopPolling did not return after cancellation")
	}
}

func TestStopPollingWithoutStartReturns(t *testing.T) {
	svc := NewService(nil, nil, nil)

	done := make(chan struct{})
	go func() {
		svc.StopPolling()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("StopPolling did not return when polling was not started")
	}
}
