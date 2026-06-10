package session_test

import (
	"context"
	"errors"
	"testing"

	"cal1604/internal/application/session"
	"cal1604/internal/device"
	"cal1604/internal/domain"
	"cal1604/internal/events"
	"cal1604/internal/infrastructure/driver"
)

// ── fakes ──

type fakeMeasureDriver struct {
	collectData []float64
	collectErr  error
	valveStatus string
	valveErr    error
	unit        string
	unitErr     error
	info        map[string]string
	infoErr     error
	resetErr    error
}

func (f *fakeMeasureDriver) Connect(_ context.Context) error    { return nil }
func (f *fakeMeasureDriver) Disconnect(_ context.Context) error { return nil }
func (f *fakeMeasureDriver) ReadValveStatus(_ context.Context) (string, error) {
	return f.valveStatus, f.valveErr
}
func (f *fakeMeasureDriver) SetValveStatus(_ context.Context, _ string) error { return nil }
func (f *fakeMeasureDriver) ReadUnit(_ context.Context) (string, error)       { return f.unit, f.unitErr }
func (f *fakeMeasureDriver) SetUnit(_ context.Context, _ string) error        { return nil }
func (f *fakeMeasureDriver) CollectData(_ context.Context, _ []int) ([]float64, error) {
	return f.collectData, f.collectErr
}
func (f *fakeMeasureDriver) ReadDeviceInfo(_ context.Context) (map[string]string, error) {
	return f.info, f.infoErr
}
func (f *fakeMeasureDriver) Reset(_ context.Context) error { return f.resetErr }
func (f *fakeMeasureDriver) CalibrateZero(_ context.Context, _ []int) ([]float64, error) {
	return []float64{0}, nil
}
func (f *fakeMeasureDriver) CalibrateFullScale(_ context.Context, _ []int, _ float64) ([]float64, error) {
	return []float64{1}, nil
}

type fakePressureDriver struct {
	pressure  float64
	pressErr  error
	stable    bool
	stableErr error
}

func (f *fakePressureDriver) Connect(_ context.Context) error                      { return nil }
func (f *fakePressureDriver) Disconnect(_ context.Context) error                   { return nil }
func (f *fakePressureDriver) SetTargetPressure(_ context.Context, _ float64) error { return nil }
func (f *fakePressureDriver) Stop(_ context.Context) error                         { return nil }
func (f *fakePressureDriver) Exhaust(_ context.Context) error                      { return nil }
func (f *fakePressureDriver) ReadCurrentPressure(_ context.Context) (float64, error) {
	return f.pressure, f.pressErr
}
func (f *fakePressureDriver) ReadUnit(_ context.Context) (string, error) { return "kPa", nil }
func (f *fakePressureDriver) SetUnit(_ context.Context, _ string) error  { return nil }
func (f *fakePressureDriver) ReadStability(_ context.Context) (bool, error) {
	return f.stable, f.stableErr
}

type fakeStore struct {
	devices map[string]domain.Device
}

func newFakeStore(devs ...domain.Device) *fakeStore {
	s := &fakeStore{devices: make(map[string]domain.Device)}
	for _, d := range devs {
		s.devices[d.ID] = d
	}
	return s
}

func (s *fakeStore) Upsert(dev domain.Device)                      { s.devices[dev.ID] = dev }
func (s *fakeStore) UpdateStatus(string, domain.DeviceStatus) bool { return true }
func (s *fakeStore) UpdateUnit(string, string) bool                { return true }
func (s *fakeStore) Delete(string)                                 {}
func (s *fakeStore) Get(id string) (domain.Device, bool)           { d, ok := s.devices[id]; return d, ok }
func (s *fakeStore) List() []domain.Device                         { return nil }
func (s *fakeStore) CheckUnitConsistency() (bool, []string)        { return true, nil }

type publisher struct {
	events []string
}

func (p *publisher) Publish(eventType string, _ any) {
	p.events = append(p.events, eventType)
}

type embedMeasure struct{ device.MeasureDriver }
type embedPressure struct{ device.PressureDriver }

type mapProvider struct {
	drivers map[string]device.ConnectionDriver
}

func (p *mapProvider) GetActiveDriver(id string) device.ConnectionDriver {
	return p.drivers[id]
}

func setupService() (*session.Service, *fakeMeasureDriver, *fakePressureDriver, *publisher) {
	mDrv := &fakeMeasureDriver{
		collectData: []float64{1.0, 2.0, 3.0},
		valveStatus: "measurement",
		unit:        "kPa",
		info:        map[string]string{"model": "WTN1604"},
	}
	pDrv := &fakePressureDriver{pressure: 100.5, stable: true}

	store := newFakeStore(
		domain.Device{ID: "m1", Type: domain.DeviceTypeMeasure, Model: "WTN1604", Host: "127.0.0.1", Port: 9000},
		domain.Device{ID: "p1", Type: domain.DeviceTypePressure, Model: "ConST811A", Host: "127.0.0.1", Port: 9001},
	)

	mp := &mapProvider{drivers: map[string]device.ConnectionDriver{
		"m1": embedMeasure{mDrv},
		"p1": embedPressure{pDrv},
	}}

	pub := &publisher{}
	svc := session.NewService(store, driver.NewFactory(), pub.Publish, mp)
	return svc, mDrv, pDrv, pub
}

// ── tests ──

func TestBindDevicesSuccess(t *testing.T) {
	svc, _, _, pub := setupService()
	token, err := svc.BindDevices("m1", "p1", "test")
	if err != nil {
		t.Fatalf("BindDevices: %v", err)
	}
	if token.BoundBy != "test" || token.MeasureDeviceID != "m1" {
		t.Fatalf("unexpected token: %+v", token)
	}
	if len(pub.events) != 1 || pub.events[0] != events.EventSessionDeviceBound {
		t.Fatalf("expected session.device_bound, got %v", pub.events)
	}
	if svc.MeasureDriver() == nil {
		t.Fatal("measure driver not bound")
	}
	if svc.PressureDriver() == nil {
		t.Fatal("pressure driver not bound")
	}
}

func TestBindDevicesMeasureNotFound(t *testing.T) {
	svc := session.NewService(newFakeStore(), driver.NewFactory(), func(string, any) {}, nil)
	_, err := svc.BindDevices("nonexistent", "p1", "test")
	if err == nil {
		t.Fatal("expected error for nonexistent device")
	}
}

func TestBindMeasureDeviceSuccess(t *testing.T) {
	svc, _, _, pub := setupService()
	token, err := svc.BindMeasureDevice("m1", "test")
	if err != nil {
		t.Fatalf("BindMeasureDevice: %v", err)
	}
	if token.BoundBy != "test" || token.MeasureDeviceID != "m1" {
		t.Fatalf("unexpected token: %+v", token)
	}
	if len(pub.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(pub.events))
	}
	if svc.MeasureDriver() == nil {
		t.Fatal("measure driver not bound")
	}
}

func TestReadPressureWithoutDevice(t *testing.T) {
	svc := session.NewService(newFakeStore(), driver.NewFactory(), func(string, any) {}, nil)
	_, err := svc.ReadPressure(context.Background(), session.BindingToken{})
	if !errors.Is(err, session.ErrBindingExpired) {
		t.Fatalf("expected ErrBindingExpired, got %v", err)
	}
}

func TestReadPressureAfterBind(t *testing.T) {
	svc, _, pDrv, _ := setupService()
	token, _ := svc.BindDevices("m1", "p1", "test")
	pDrv.pressure = 200.5
	val, err := svc.ReadPressure(context.Background(), token)
	if err != nil {
		t.Fatalf("ReadPressure: %v", err)
	}
	if val != 200.5 {
		t.Fatalf("expected 200.5, got %f", val)
	}
}

func TestReadStabilityAfterBind(t *testing.T) {
	svc, _, pDrv, _ := setupService()
	token, _ := svc.BindDevices("m1", "p1", "test")
	pDrv.stable = false
	val, err := svc.ReadStability(context.Background(), token)
	if err != nil {
		t.Fatalf("ReadStability: %v", err)
	}
	if val {
		t.Fatal("expected false")
	}
}

func TestReadMeasureDataWithoutDevice(t *testing.T) {
	svc := session.NewService(newFakeStore(), driver.NewFactory(), func(string, any) {}, nil)
	_, err := svc.ReadMeasureData(context.Background(), session.BindingToken{})
	if !errors.Is(err, session.ErrBindingExpired) {
		t.Fatalf("expected ErrBindingExpired, got %v", err)
	}
}

func TestReadMeasureDataAfterBind(t *testing.T) {
	svc, mDrv, _, _ := setupService()
	token, _ := svc.BindMeasureDevice("m1", "test")
	mDrv.collectData = []float64{10.1, 20.2, 30.3}
	data, err := svc.ReadMeasureData(context.Background(), token)
	if err != nil {
		t.Fatalf("ReadMeasureData: %v", err)
	}
	if len(data) != 3 || data[0] != 10.1 {
		t.Fatalf("unexpected data: %v", data)
	}
}

func TestReadValveStatus(t *testing.T) {
	svc, mDrv, _, _ := setupService()
	token, _ := svc.BindMeasureDevice("m1", "test")
	mDrv.valveStatus = "calibration"
	val, err := svc.ReadValveStatus(context.Background(), token)
	if err != nil {
		t.Fatalf("ReadValveStatus: %v", err)
	}
	if val != "calibration" {
		t.Fatalf("expected calibration, got %s", val)
	}
}

func TestReadMeasureUnit(t *testing.T) {
	svc, mDrv, _, _ := setupService()
	token, _ := svc.BindMeasureDevice("m1", "test")
	mDrv.unit = "MPa"
	val, err := svc.ReadMeasureUnit(context.Background(), token)
	if err != nil {
		t.Fatalf("ReadMeasureUnit: %v", err)
	}
	if val != "MPa" {
		t.Fatalf("expected MPa, got %s", val)
	}
}

func TestReadDeviceInfo(t *testing.T) {
	svc, mDrv, _, _ := setupService()
	token, _ := svc.BindMeasureDevice("m1", "test")
	mDrv.info = map[string]string{"model": "WTN1604", "version": "2.0"}
	info, err := svc.ReadDeviceInfo(context.Background(), token)
	if err != nil {
		t.Fatalf("ReadDeviceInfo: %v", err)
	}
	if info["model"] != "WTN1604" {
		t.Fatalf("unexpected info: %v", info)
	}
}

func TestResetDevice(t *testing.T) {
	svc, mDrv, _, _ := setupService()
	token, _ := svc.BindMeasureDevice("m1", "test")
	if err := svc.ResetDevice(context.Background(), token); err != nil {
		t.Fatalf("ResetDevice: %v", err)
	}
	mDrv.resetErr = errors.New("reset failed")
	if err := svc.ResetDevice(context.Background(), token); err == nil {
		t.Fatal("expected reset error")
	}
}

func TestBindDevicesOnlyMeasure(t *testing.T) {
	svc, _, _, _ := setupService()
	token, err := svc.BindDevices("m1", "", "test")
	if err != nil {
		t.Fatalf("BindDevices with empty pressure: %v", err)
	}
	if token.PressureDeviceID != "" {
		t.Fatalf("expected empty pressure device id, got %q", token.PressureDeviceID)
	}
	if svc.MeasureDriver() == nil {
		t.Fatal("measure driver not bound")
	}
	if svc.PressureDriver() != nil {
		t.Fatal("expected nil pressure driver")
	}
}
