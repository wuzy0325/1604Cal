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

type fakePressureDriver struct {
	targets          []float64
	currentPressure  float64
	stable           bool
	startControlCall int
	stopCalled       bool
}

func (f *fakePressureDriver) Connect(_ context.Context) error    { return nil }
func (f *fakePressureDriver) Disconnect(_ context.Context) error { return nil }
func (f *fakePressureDriver) SetTargetPressure(_ context.Context, target float64) error {
	f.targets = append(f.targets, target)
	f.currentPressure = target
	return nil
}
func (f *fakePressureDriver) Stop(_ context.Context) error {
	f.stopCalled = true
	return nil
}
func (f *fakePressureDriver) Exhaust(_ context.Context) error { return nil }
func (f *fakePressureDriver) ReadCurrentPressure(_ context.Context) (float64, error) {
	return f.currentPressure, nil
}
func (f *fakePressureDriver) ReadUnit(_ context.Context) (string, error) { return "kPa", nil }
func (f *fakePressureDriver) SetUnit(_ context.Context, _ string) error  { return nil }
func (f *fakePressureDriver) ReadStability(_ context.Context) (bool, error) {
	return f.stable, nil
}
func (f *fakePressureDriver) StartControl(_ context.Context) error {
	f.startControlCall++
	return nil
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

func setupMeasurementServiceWithPressure() (*measurement.Service, *fakeMeasureDriver, *fakePressureDriver) {
	mDrv := &fakeMeasureDriver{data: []float64{1.1, 2.2, 3.3}}
	pDrv := &fakePressureDriver{stable: true}
	store := &fakeStore{devices: map[string]domain.Device{
		"m1": {ID: "m1", Type: domain.DeviceTypeMeasure, Model: "WTN1604", Host: "127.0.0.1", Port: 9000},
		"p1": {ID: "p1", Type: domain.DeviceTypePressure, Model: "ConST811A", Host: "127.0.0.1", Port: 9001},
	}}
	sessSvc := session.NewService(store, driver.NewFactory(), func(string, any) {}, &mapProvider{
		drivers: map[string]device.ConnectionDriver{
			"m1": embedMD{mDrv},
			"p1": embedPD{pDrv},
		},
	})
	_ = sessSvc.BindDevices("m1", "p1")
	svc := measurement.NewService(sessSvc, func(string, any) {})
	return svc, mDrv, pDrv
}

type embedPD struct{ device.PressureDriver }

func TestInitialState(t *testing.T) {
	svc, _ := setupMeasurementService()
	if svc.State() != measurement.StateIdle {
		t.Fatalf("expected idle, got %s", svc.State())
	}
}

func TestGeneratePressurePointsUsesMeasurementConfig(t *testing.T) {
	svc, _ := setupMeasurementService()
	svc.SetConfig(measurement.Config{
		MinPressure:  0,
		MaxPressure:  100,
		PointCount:   5,
		Precision:    2,
		PressureMode: "roundTrip",
	})

	points, err := svc.GeneratePressurePoints()
	if err != nil {
		t.Fatalf("GeneratePressurePoints: %v", err)
	}

	if len(points) != 9 {
		t.Fatalf("expected 9 points (5 forward + 4 backward), got %d", len(points))
	}

	if points[0].Direction != "forward" {
		t.Fatalf("expected first point direction forward, got %s", points[0].Direction)
	}

	if points[len(points)-1].Direction != "backward" {
		t.Fatalf("expected last point direction backward, got %s", points[len(points)-1].Direction)
	}

	// forward: 0, 25, 50, 75, 100; backward: 100, 75, 50, 25
	expected := []float64{0, 25, 50, 75, 100, 100, 75, 50, 25}
	for i, exp := range expected {
		if points[i].TargetPressure != exp {
			t.Fatalf("points[%d].TargetPressure = %v, want %v", i, points[i].TargetPressure, exp)
		}
	}
}

func TestStartWorkflowUsesGeneratedPointsAndTransitionsToReady(t *testing.T) {
	svc, _ := setupMeasurementService()
	svc.SetConfig(measurement.Config{
		MinPressure: 0,
		MaxPressure: 20,
		PointCount:  3,
		Precision:   2,
	})
	if _, err := svc.GeneratePressurePoints(); err != nil {
		t.Fatalf("GeneratePressurePoints: %v", err)
	}

	if err := svc.StartWorkflow(context.Background(), []int{1, 2}); err != nil {
		t.Fatalf("StartWorkflow: %v", err)
	}

	if svc.State() != measurement.StateReady {
		t.Fatalf("expected ready, got %s", svc.State())
	}

	session := svc.GetSession()
	if session == nil {
		t.Fatal("expected measurement session to be created")
	}
	if len(session.Points) != 3 {
		t.Fatalf("expected 3 session points, got %d", len(session.Points))
	}
	if len(session.Config.Channels) != 2 {
		t.Fatalf("expected workflow channels to be stored, got %+v", session.Config.Channels)
	}
}

func TestSetConfigInvalidatesExistingPoints(t *testing.T) {
	svc, _ := setupMeasurementService()
	svc.SetConfig(measurement.Config{
		MinPressure: 0,
		MaxPressure: 20,
		PointCount:  3,
		Precision:   2,
	})
	if _, err := svc.GeneratePressurePoints(); err != nil {
		t.Fatalf("GeneratePressurePoints: %v", err)
	}

	svc.SetConfig(measurement.Config{
		MinPressure: 0,
		MaxPressure: 50,
		PointCount:  5,
		Precision:   2,
	})

	if err := svc.StartWorkflow(context.Background(), []int{1}); err == nil {
		t.Fatal("expected workflow start to require regenerated points after config change")
	}
}

func TestStopWorkflowUpdatesSessionStatusToStopped(t *testing.T) {
	svc, _ := setupMeasurementService()
	svc.SetConfig(measurement.Config{
		MinPressure: 0,
		MaxPressure: 10,
		PointCount:  2,
		Precision:   2,
	})
	if _, err := svc.GeneratePressurePoints(); err != nil {
		t.Fatalf("GeneratePressurePoints: %v", err)
	}
	if err := svc.StartWorkflow(context.Background(), []int{1}); err != nil {
		t.Fatalf("StartWorkflow: %v", err)
	}

	if err := svc.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	session := svc.GetSession()
	if session == nil {
		t.Fatal("expected measurement session after stop")
	}
	if session.Status != measurement.StateStopped {
		t.Fatalf("expected session status stopped, got %s", session.Status)
	}
	if session.EndTime == nil {
		t.Fatal("expected session end time to be recorded")
	}
}

func TestResumeRealtimeSamplingSyncsWorkflowSessionStatus(t *testing.T) {
	svc, _ := setupMeasurementService()
	svc.SetConfig(measurement.Config{
		MinPressure: 0,
		MaxPressure: 10,
		PointCount:  2,
		Precision:   2,
	})
	if _, err := svc.GeneratePressurePoints(); err != nil {
		t.Fatalf("GeneratePressurePoints: %v", err)
	}
	if err := svc.StartWorkflow(context.Background(), []int{1}); err != nil {
		t.Fatalf("StartWorkflow: %v", err)
	}
	if err := svc.Pause(); err != nil {
		t.Fatalf("Pause: %v", err)
	}
	if err := svc.Start(context.Background(), []int{1}); err != nil {
		t.Fatalf("Start realtime resume: %v", err)
	}

	session := svc.GetSession()
	if session == nil {
		t.Fatal("expected measurement session after resume")
	}
	if session.Status != measurement.StateCollecting {
		t.Fatalf("expected session status collecting after resume, got %s", session.Status)
	}
	_ = svc.Stop()
}

func TestRunAutoCollectionAdvancesMeasurementPoints(t *testing.T) {
	svc, _, pressureDrv := setupMeasurementServiceWithPressure()
	svc.SetConfig(measurement.Config{
		MinPressure:  0,
		MaxPressure:  10,
		PointCount:   2,
		Precision:    2,
		AverageCount: 1,
		StableWaitMs: 10,
		ControlMode:  "auto",
	})
	if _, err := svc.GeneratePressurePoints(); err != nil {
		t.Fatalf("GeneratePressurePoints: %v", err)
	}
	if err := svc.StartWorkflow(context.Background(), []int{1}); err != nil {
		t.Fatalf("StartWorkflow: %v", err)
	}

	if err := svc.RunAutoCollection(context.Background()); err != nil {
		t.Fatalf("RunAutoCollection: %v", err)
	}

	points := svc.GetPoints()
	if len(points) != 2 {
		t.Fatalf("expected 2 points, got %d", len(points))
	}
	if points[0].Status != "completed" || points[1].Status != "completed" {
		t.Fatalf("expected all points completed, got %+v", points)
	}
	if len(pressureDrv.targets) != 2 {
		t.Fatalf("expected 2 pressure targets, got %+v", pressureDrv.targets)
	}
}

func TestManualCollectCapturesPointData(t *testing.T) {
	svc, _, pressureDrv := setupMeasurementServiceWithPressure()
	svc.SetConfig(measurement.Config{
		MinPressure:  0,
		MaxPressure:  20,
		PointCount:   3,
		Precision:    2,
		AverageCount: 1,
		StableWaitMs: 10,
		ControlMode:  "manual",
	})
	if _, err := svc.GeneratePressurePoints(); err != nil {
		t.Fatalf("GeneratePressurePoints: %v", err)
	}
	if err := svc.StartWorkflow(context.Background(), []int{1, 2}); err != nil {
		t.Fatalf("StartWorkflow: %v", err)
	}

	if err := svc.ManualPressurize(context.Background(), 1); err != nil {
		t.Fatalf("ManualPressurize: %v", err)
	}
	if err := svc.ManualCollect(context.Background(), 1); err != nil {
		t.Fatalf("ManualCollect: %v", err)
	}

	points := svc.GetPoints()
	if points[0].Status != "completed" {
		t.Fatalf("expected first point completed, got %+v", points[0])
	}
	if len(points[0].CollectedData) == 0 {
		t.Fatalf("expected collected data on first point, got %+v", points[0])
	}
	if len(pressureDrv.targets) != 1 {
		t.Fatalf("expected 1 pressure target, got %+v", pressureDrv.targets)
	}
}

func TestSetStateTransitions(t *testing.T) {
	tests := []struct {
		name      string
		from      measurement.State
		to        measurement.State
		wantError bool
	}{
		{name: "idle_to_pressurizing", from: measurement.StateIdle, to: measurement.StatePressurizing},
		{name: "idle_to_collecting", from: measurement.StateIdle, to: measurement.StateCollecting},
		{name: "pressurizing_to_stabilizing", from: measurement.StatePressurizing, to: measurement.StateStabilizing},
		{name: "pressurizing_to_paused", from: measurement.StatePressurizing, to: measurement.StatePaused},
		{name: "stabilizing_to_collecting", from: measurement.StateStabilizing, to: measurement.StateCollecting},
		{name: "collecting_to_completed", from: measurement.StateCollecting, to: measurement.StateCompleted},
		{name: "completed_to_idle", from: measurement.StateCompleted, to: measurement.StateIdle},
		{name: "error_to_idle", from: measurement.StateError, to: measurement.StateIdle},
		{name: "paused_to_collecting", from: measurement.StatePaused, to: measurement.StateCollecting},
		{name: "paused_to_pressurizing", from: measurement.StatePaused, to: measurement.StatePressurizing},
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
		measurement.StatePressurizing: {measurement.StatePressurizing},
		measurement.StateStabilizing: {measurement.StatePressurizing, measurement.StateStabilizing},
		measurement.StateCollecting:  {measurement.StatePressurizing, measurement.StateStabilizing, measurement.StateCollecting},
		measurement.StateCompleted:   {measurement.StatePressurizing, measurement.StateStabilizing, measurement.StateCollecting, measurement.StateCompleted},
		measurement.StateError:       {measurement.StatePressurizing, measurement.StateError},
		measurement.StatePaused:      {measurement.StatePressurizing, measurement.StatePaused},
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
		})
	}
}

func TestMeasurementAlarmConfig(t *testing.T) {
	svc, _ := setupMeasurementService()
	svc.SetAlarmConfig(measurement.AlarmConfig{
		Enabled:         true,
		EnabledChannels: []int{1, 2},
		ConfirmOnAlarm:  true,
		SoundEnabled:    false,
		Threshold:       0.01,
	})

	cfg := svc.GetAlarmConfig()
	if !cfg.Enabled {
		t.Fatal("expected alarm enabled")
	}
	if len(cfg.EnabledChannels) != 2 {
		t.Fatalf("expected 2 enabled channels, got %d", len(cfg.EnabledChannels))
	}
	if !cfg.ConfirmOnAlarm {
		t.Fatal("expected confirm on alarm")
	}
}

func TestMeasurementAlarmCheckNoAlarm(t *testing.T) {
	svc, _ := setupMeasurementService()
	svc.SetAlarmConfig(measurement.AlarmConfig{
		Enabled:         true,
		EnabledChannels: []int{1},
		ConfirmOnAlarm:  true,
		Threshold:       0.1,
	})

	point := measurement.Point{
		Index:          1,
		TargetPressure: 100,
		ActualPressure: 100.05,
	}

	alarm, err := svc.CheckAlarm(point)
	if err != nil {
		t.Fatalf("CheckAlarm: %v", err)
	}
	if alarm != nil {
		t.Fatalf("expected no alarm, got %+v", alarm)
	}
}

func TestMeasurementAlarmCheckTriggersAlarm(t *testing.T) {
	svc, _ := setupMeasurementService()
	svc.SetAlarmConfig(measurement.AlarmConfig{
		Enabled:         true,
		EnabledChannels: []int{1},
		ConfirmOnAlarm:  true,
		Threshold:       0.0004,
	})

	point := measurement.Point{
		Index:          1,
		TargetPressure: 100,
		ActualPressure: 100.05,
		CollectedData:  []float64{100.05},
	}

	alarm, err := svc.CheckAlarm(point)
	if err != nil {
		t.Fatalf("CheckAlarm: %v", err)
	}
	if alarm == nil {
		t.Fatal("expected alarm to be triggered")
	}
	if len(alarm.OverLimitChannels) == 0 {
		t.Fatal("expected over limit channels")
	}
}

func TestMeasurementAlarmBlocksWhenConfirmRequired(t *testing.T) {
	svc, _ := setupMeasurementService()
	svc.SetAlarmConfig(measurement.AlarmConfig{
		Enabled:         true,
		EnabledChannels: []int{1},
		ConfirmOnAlarm:  true,
		Threshold:       0.0004,
	})

	point := measurement.Point{
		Index:          1,
		TargetPressure: 100,
		ActualPressure: 100.05,
		CollectedData:  []float64{100.05},
	}

	alarm, err := svc.CheckAlarm(point)
	if err != nil {
		t.Fatalf("CheckAlarm: %v", err)
	}
	if alarm == nil {
		t.Fatal("expected alarm")
	}

	if !svc.IsAlarmPending() {
		t.Fatal("expected alarm to be pending")
	}

	err = svc.ResolveAlarm("continue")
	if err != nil {
		t.Fatalf("ResolveAlarm: %v", err)
	}

	if svc.IsAlarmPending() {
		t.Fatal("expected alarm to be resolved")
	}
}
