package measurement_test

import (
	"bytes"
	"context"
	"testing"

	"cal1604/internal/application/measurement"
	"cal1604/internal/application/session"
	"cal1604/internal/device"
	"cal1604/internal/domain"
	"cal1604/internal/infrastructure/driver"
)

// fakeMeasureDriver 最小实现，仅 CollectData。
type fakeMeasureDriver struct {
	data []float64
	err  error
}

func (f *fakeMeasureDriver) Connect(_ context.Context) error    { return nil }
func (f *fakeMeasureDriver) Disconnect(_ context.Context) error { return nil }
func (f *fakeMeasureDriver) ReadValveStatus(_ context.Context) (string, error) {
	return "measurement", nil
}
func (f *fakeMeasureDriver) SetValveStatus(_ context.Context, _ string) error { return nil }
func (f *fakeMeasureDriver) ReadUnit(_ context.Context) (string, error)       { return "kPa", nil }
func (f *fakeMeasureDriver) SetUnit(_ context.Context, _ string) error        { return nil }
func (f *fakeMeasureDriver) CollectData(_ context.Context, _ []int) ([]float64, error) {
	return f.data, f.err
}
func (f *fakeMeasureDriver) ReadDeviceInfo(_ context.Context) (map[string]string, error) {
	return nil, nil
}
func (f *fakeMeasureDriver) Reset(_ context.Context) error { return nil }
func (f *fakeMeasureDriver) CalibrateZero(_ context.Context, _ []int) ([]float64, error) {
	return nil, nil
}
func (f *fakeMeasureDriver) CalibrateFullScale(_ context.Context, _ []int, _ float64) ([]float64, error) {
	return nil, nil
}

type fakeStore struct {
	devices map[string]domain.Device
}

func (s *fakeStore) Upsert(dev domain.Device)                      { s.devices[dev.ID] = dev }
func (s *fakeStore) UpdateStatus(string, domain.DeviceStatus) bool { return true }
func (s *fakeStore) Delete(string)                                 {}
func (s *fakeStore) Get(id string) (domain.Device, bool)           { d, ok := s.devices[id]; return d, ok }
func (s *fakeStore) List() []domain.Device                         { return nil }
func (s *fakeStore) CheckUnitConsistency() (bool, []string)        { return true, nil }

type embedMD struct{ device.MeasureDriver }

type mapProvider struct {
	drivers map[string]device.ConnectionDriver
}

func (p *mapProvider) GetActiveDriver(id string) device.ConnectionDriver {
	return p.drivers[id]
}

func setupMeasurementService() (*measurement.Service, *fakeMeasureDriver) {
	mDrv := &fakeMeasureDriver{data: []float64{1.1, 2.2, 3.3}}
	store := &fakeStore{devices: map[string]domain.Device{
		"m1": {ID: "m1", Type: domain.DeviceTypeMeasure, Model: "WTN1604", Host: "127.0.0.1", Port: 9000},
	}}
	sessSvc := session.NewService(store, driver.NewFactory(), func(string, any) {}, &mapProvider{
		drivers: map[string]device.ConnectionDriver{"m1": embedMD{mDrv}},
	})
	_ = sessSvc.BindMeasureDevice("m1")
	svc := measurement.NewService(sessSvc, func(string, any) {})
	return svc, mDrv
}

func TestInitialState(t *testing.T) {
	svc, _ := setupMeasurementService()
	if svc.State() != measurement.StateIdle {
		t.Fatalf("expected idle, got %s", svc.State())
	}
}

func TestSetStateTransitions(t *testing.T) {
	tests := []struct {
		name      string
		from      measurement.State
		to        measurement.State
		wantError bool
	}{
		{name: "idle_to_pressuring", from: measurement.StateIdle, to: measurement.StatePressuring},
		{name: "idle_to_collecting_invalid", from: measurement.StateIdle, to: measurement.StateCollecting, wantError: true},
		{name: "pressuring_to_stabilizing", from: measurement.StatePressuring, to: measurement.StateStabilizing},
		{name: "pressuring_to_paused", from: measurement.StatePressuring, to: measurement.StatePaused},
		{name: "stabilizing_to_collecting", from: measurement.StateStabilizing, to: measurement.StateCollecting},
		{name: "collecting_to_completed", from: measurement.StateCollecting, to: measurement.StateCompleted},
		{name: "completed_to_idle", from: measurement.StateCompleted, to: measurement.StateIdle},
		{name: "error_to_idle", from: measurement.StateError, to: measurement.StateIdle},
		{name: "paused_to_collecting", from: measurement.StatePaused, to: measurement.StateCollecting},
		{name: "paused_to_pressuring", from: measurement.StatePaused, to: measurement.StatePressuring},
		{name: "paused_to_idle", from: measurement.StatePaused, to: measurement.StateIdle},
		{name: "collecting_to_idle_invalid", from: measurement.StateCollecting, to: measurement.StateIdle, wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _ := setupMeasurementService()
			reachState(t, svc, tt.from)

			err := svc.SetState(tt.to)
			if tt.wantError {
				if err == nil {
					t.Fatalf("expected transition %s -> %s to fail", tt.from, tt.to)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected transition error %s -> %s: %v", tt.from, tt.to, err)
			}
			if got := svc.State(); got != tt.to {
				t.Fatalf("expected state %s, got %s", tt.to, got)
			}
		})
	}
}

func reachState(t *testing.T, svc *measurement.Service, target measurement.State) {
	t.Helper()

	if target == measurement.StateIdle {
		return
	}

	steps := map[measurement.State][]measurement.State{
		measurement.StatePressuring:  {measurement.StatePressuring},
		measurement.StateStabilizing: {measurement.StatePressuring, measurement.StateStabilizing},
		measurement.StateCollecting:  {measurement.StatePressuring, measurement.StateStabilizing, measurement.StateCollecting},
		measurement.StateCompleted:   {measurement.StatePressuring, measurement.StateStabilizing, measurement.StateCollecting, measurement.StateCompleted},
		measurement.StateError:       {measurement.StatePressuring, measurement.StateError},
		measurement.StatePaused:      {measurement.StatePressuring, measurement.StatePaused},
	}

	path, ok := steps[target]
	if !ok {
		t.Fatalf("unsupported target state in test helper: %s", target)
	}

	for _, state := range path {
		if err := svc.SetState(state); err != nil {
			t.Fatalf("reach state %s failed at %s: %v", target, state, err)
		}
	}
}

func TestStartTransition(t *testing.T) {
	svc, _ := setupMeasurementService()
	if err := svc.Start(context.Background(), []int{1, 2, 3}); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if svc.State() != measurement.StateCollecting {
		t.Fatalf("expected collecting, got %s", svc.State())
	}
	_ = svc.Stop()
}

func TestStartWithoutDevice(t *testing.T) {
	sessSvc := session.NewService(&fakeStore{}, driver.NewFactory(), func(string, any) {}, nil)
	svc := measurement.NewService(sessSvc, func(string, any) {})
	if err := svc.Start(context.Background(), []int{1}); err == nil {
		t.Fatal("expected error when no device bound")
	}
}

func TestPauseFromCollecting(t *testing.T) {
	svc, _ := setupMeasurementService()
	_ = svc.Start(context.Background(), []int{1, 2})
	defer svc.Stop()
	if err := svc.Pause(); err != nil {
		t.Fatalf("Pause: %v", err)
	}
	if svc.State() != measurement.StatePaused {
		t.Fatalf("expected paused, got %s", svc.State())
	}
}

func TestPauseFromIdleFails(t *testing.T) {
	svc, _ := setupMeasurementService()
	if err := svc.Pause(); err == nil {
		t.Fatal("expected error pausing from idle")
	}
}

func TestStopFromCollecting(t *testing.T) {
	svc, _ := setupMeasurementService()
	_ = svc.Start(context.Background(), []int{1})
	if err := svc.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if svc.State() != measurement.StateIdle {
		t.Fatalf("expected idle, got %s", svc.State())
	}
}

func TestStopFromPaused(t *testing.T) {
	svc, _ := setupMeasurementService()
	_ = svc.Start(context.Background(), []int{1})
	_ = svc.Pause()
	if err := svc.Stop(); err != nil {
		t.Fatalf("Stop from paused: %v", err)
	}
	if svc.State() != measurement.StateIdle {
		t.Fatalf("expected idle, got %s", svc.State())
	}
}

func TestStopFromIdleFails(t *testing.T) {
	svc, _ := setupMeasurementService()
	if err := svc.Stop(); err == nil {
		t.Fatal("expected error stopping from idle")
	}
}

func TestResumeFromPaused(t *testing.T) {
	svc, _ := setupMeasurementService()
	_ = svc.Start(context.Background(), []int{1, 2})
	_ = svc.Pause()
	if err := svc.Start(context.Background(), []int{1, 2}); err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if svc.State() != measurement.StateCollecting {
		t.Fatalf("expected collecting, got %s", svc.State())
	}
	_ = svc.Stop()
}

func TestWriteCSVEmpty(t *testing.T) {
	svc, _ := setupMeasurementService()
	var buf bytes.Buffer
	if err := svc.WriteCSV(&buf); err == nil {
		t.Fatal("expected error for empty data")
	}
}

func TestStateTransitions(t *testing.T) {
	tests := []struct {
		name    string
		from    measurement.State
		to      measurement.State
		wantErr bool
	}{
		{"idle_to_collecting", measurement.StateIdle, measurement.StateCollecting, false},
		{"idle_to_paused", measurement.StateIdle, measurement.StatePaused, true},
		{"collecting_to_paused", measurement.StateCollecting, measurement.StatePaused, false},
		{"collecting_to_idle", measurement.StateCollecting, measurement.StateIdle, false},
		{"collecting_to_collecting", measurement.StateCollecting, measurement.StateCollecting, true},
		{"paused_to_collecting", measurement.StatePaused, measurement.StateCollecting, false},
		{"paused_to_idle", measurement.StatePaused, measurement.StateIdle, false},
		{"paused_to_paused", measurement.StatePaused, measurement.StatePaused, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _ := setupMeasurementService()
			switch tt.from {
			case measurement.StateCollecting:
				_ = svc.Start(context.Background(), []int{1})
			case measurement.StatePaused:
				_ = svc.Start(context.Background(), []int{1})
				_ = svc.Pause()
			}

			var err error
			switch tt.to {
			case measurement.StateCollecting:
				err = svc.Start(context.Background(), []int{1})
			case measurement.StatePaused:
				err = svc.Pause()
			case measurement.StateIdle:
				err = svc.Stop()
			}

			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			_ = svc.Stop()
		})
	}
}
