package calibration

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cal1604/internal/domain"
	"cal1604/internal/report"
	"cal1604/internal/workflow"
)

type valveGateFakeMeasureDriver struct {
	valveStatus string
}

func (d *valveGateFakeMeasureDriver) Connect(_ context.Context) error { return nil }

func (d *valveGateFakeMeasureDriver) Disconnect(_ context.Context) error { return nil }

func (d *valveGateFakeMeasureDriver) ReadValveStatus(_ context.Context) (string, error) {
	if d.valveStatus == "" {
		return "measurement", nil
	}
	return d.valveStatus, nil
}

func (d *valveGateFakeMeasureDriver) SetValveStatus(_ context.Context, status string) error {
	d.valveStatus = status
	return nil
}

func (d *valveGateFakeMeasureDriver) ReadUnit(_ context.Context) (string, error) { return "kPa", nil }

func (d *valveGateFakeMeasureDriver) SetUnit(_ context.Context, _ string) error { return nil }

func (d *valveGateFakeMeasureDriver) CollectData(_ context.Context, channels []int) ([]float64, error) {
	result := make([]float64, len(channels))
	return result, nil
}

func (d *valveGateFakeMeasureDriver) ReadDeviceInfo(_ context.Context) (map[string]string, error) {
	return map[string]string{"model": "fake"}, nil
}

func (d *valveGateFakeMeasureDriver) Reset(_ context.Context) error { return nil }

func newValveGateTestService() *Service {
	svc := NewService(workflow.NewSessionMachine(), nil, nil, nil, nil, nil)
	svc.measureDriver = &valveGateFakeMeasureDriver{valveStatus: "measurement"}
	svc.config = CalibrationConfig{
		Channels:       []int{1},
		PressurePoints: 2,
		ControlMode:    "manual",
	}
	return svc
}

func TestValidateStartPrerequisitesValveGateSwitch(t *testing.T) {
	svc := newValveGateTestService()
	svc.SetStartPrerequisiteConfig(StartPrerequisiteConfig{EnforceValveCalibration: true})

	err := svc.ValidateStartPrerequisites(context.Background())
	if err == nil || !strings.Contains(err.Error(), "valve must be in calibration state") {
		t.Fatalf("expected valve gate error, got %v", err)
	}

	svc.SetStartPrerequisiteConfig(StartPrerequisiteConfig{EnforceValveCalibration: false})
	if err := svc.ValidateStartPrerequisites(context.Background()); err != nil {
		t.Fatalf("expected prerequisites pass when valve gate disabled, got %v", err)
	}
}

func TestStartCalibrationValveGateSwitch(t *testing.T) {
	svc := newValveGateTestService()
	svc.SetStartPrerequisiteConfig(StartPrerequisiteConfig{EnforceValveCalibration: true})

	err := svc.StartCalibration(context.Background())
	if err == nil || !strings.Contains(err.Error(), "valve must be in calibration state") {
		t.Fatalf("expected valve gate error, got %v", err)
	}

	svc.SetStartPrerequisiteConfig(StartPrerequisiteConfig{EnforceValveCalibration: false})
	if err := svc.StartCalibration(context.Background()); err != nil {
		t.Fatalf("expected start pass when valve gate disabled, got %v", err)
	}
}

type calibrationCollectFakeMeasureDriver struct {
	collectDataCalled             bool
	collectCalibrationPointCalled bool
	collectCalibrationPointCount  int
}

func (d *calibrationCollectFakeMeasureDriver) Connect(_ context.Context) error { return nil }

func (d *calibrationCollectFakeMeasureDriver) Disconnect(_ context.Context) error { return nil }

func (d *calibrationCollectFakeMeasureDriver) ReadValveStatus(_ context.Context) (string, error) {
	return "calibration", nil
}

func (d *calibrationCollectFakeMeasureDriver) SetValveStatus(_ context.Context, _ string) error {
	return nil
}

func (d *calibrationCollectFakeMeasureDriver) ReadUnit(_ context.Context) (string, error) {
	return "kPa", nil
}

func (d *calibrationCollectFakeMeasureDriver) SetUnit(_ context.Context, _ string) error { return nil }

func (d *calibrationCollectFakeMeasureDriver) CollectData(_ context.Context, _ []int) ([]float64, error) {
	d.collectDataCalled = true
	return nil, errors.New("collect data command not allowed in calibration mode")
}

func (d *calibrationCollectFakeMeasureDriver) CollectCalibrationPoint(_ context.Context, pointIndex int, targetPressure float64) ([]float64, error) {
	d.collectCalibrationPointCalled = true
	d.collectCalibrationPointCount++
	if pointIndex != 1 {
		return nil, errors.New("unexpected point index")
	}
	if targetPressure != 10 {
		return nil, errors.New("unexpected target pressure")
	}
	return []float64{10.11, 10.22}, nil
}

func (d *calibrationCollectFakeMeasureDriver) ReadDeviceInfo(_ context.Context) (map[string]string, error) {
	return map[string]string{"model": "fake"}, nil
}

func (d *calibrationCollectFakeMeasureDriver) Reset(_ context.Context) error { return nil }

func TestCollectUsesCalibrationPointCommandWhenSupported(t *testing.T) {
	drv := &calibrationCollectFakeMeasureDriver{}
	svc := NewService(workflow.NewSessionMachine(), nil, nil, nil, nil, nil)
	svc.measureDriver = drv
	svc.config = CalibrationConfig{
		Channels:       []int{1, 2},
		AverageCount:   5,
		PressurePoints: 2,
	}
	svc.pressurePoints = []PressurePoint{
		{
			Index:          1,
			TargetPressure: 10,
			Status:         "stabilizing",
		},
	}

	data, err := svc.Collect(context.Background(), 1)
	if err != nil {
		t.Fatalf("Collect should succeed with calibration point command, got %v", err)
	}

	if !drv.collectCalibrationPointCalled {
		t.Fatalf("expected CollectCalibrationPoint to be called")
	}
	if drv.collectDataCalled {
		t.Fatalf("expected CollectData not to be called when calibration command is supported")
	}
	if drv.collectCalibrationPointCount != 5 {
		t.Fatalf("expected CollectCalibrationPoint to be called 5 times, got %d", drv.collectCalibrationPointCount)
	}

	if len(data) != 2 || data[0] != 10.11 || data[1] != 10.22 {
		t.Fatalf("unexpected collect result: %#v", data)
	}

	if svc.pressurePoints[0].Status != "completed" {
		t.Fatalf("expected point status completed, got %s", svc.pressurePoints[0].Status)
	}
}

func TestRetryPointManualModeWithoutPressureDeviceResetsOnly(t *testing.T) {
	svc := NewService(workflow.NewSessionMachine(), nil, nil, nil, nil, nil)
	svc.config = CalibrationConfig{ControlMode: "manual"}
	svc.pressurePoints = []PressurePoint{
		{
			Index:          1,
			TargetPressure: 10,
			Status:         "completed",
			CollectedData:  []float64{10.12, 10.23},
			ActualPressure: 10.15,
		},
	}

	if err := svc.RetryPoint(context.Background(), 1); err != nil {
		t.Fatalf("RetryPoint should not fail in manual mode without pressure device, got %v", err)
	}

	point := svc.pressurePoints[0]
	if point.Status != "pending" {
		t.Fatalf("expected point status pending after retry reset, got %s", point.Status)
	}
	if point.CollectedData != nil {
		t.Fatalf("expected collected data to be cleared, got %#v", point.CollectedData)
	}
	if point.ActualPressure != 0 {
		t.Fatalf("expected actual pressure reset to 0, got %v", point.ActualPressure)
	}
}
