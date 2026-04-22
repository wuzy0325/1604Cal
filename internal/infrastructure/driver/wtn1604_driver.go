package driver

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// WTN1604 计量设备驱动
// ---------------------------------------------------------------------------

type WTN1604Driver struct {
	base *tcpConnectionDriver
}

func newWTN1604Driver(host string, port int) *WTN1604Driver {
	return &WTN1604Driver{base: newTCPConnectionDriver("WTN1604", host, port)}
}

func (d *WTN1604Driver) Connect(ctx context.Context) error    { return d.base.Connect(ctx) }
func (d *WTN1604Driver) Disconnect(ctx context.Context) error { return d.base.Disconnect(ctx) }

func (d *WTN1604Driver) ReadValveStatus(ctx context.Context) (string, error) {
	resp, err := d.base.sendWTN1604Command(ctx, "@01  0", 3*time.Second)
	if err != nil {
		return "", fmt.Errorf("read valve status: %w", err)
	}

	raw := strings.TrimSpace(resp)
	val := strings.TrimSpace(strings.TrimPrefix(raw, "A"))
	if val == "" {
		val = raw
	}

	if num, parseErr := strconv.Atoi(strings.TrimSpace(val)); parseErr == nil {
		switch num {
		case 1:
			return "calibration", nil
		case 0, 2, 3:
			// 现场兼容：部分 1604 固件在 RUN/测量态返回 3，按 measurement 处理。
			return "measurement", nil
		}
	}

	switch strings.ToLower(strings.TrimSpace(val)) {
	case "calibration", "calibrate", "open", "opened", "on":
		return "calibration", nil
	case "measurement", "measure", "close", "closed", "off":
		return "measurement", nil
	default:
		return raw, nil
	}
}

func (d *WTN1604Driver) SetValveStatus(ctx context.Context, status string) error {
	var cmd string
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "calibration", "calibrate", "1":
		cmd = "w0C01"
	case "measurement", "measure", "2":
		cmd = "w0C00"
	default:
		return fmt.Errorf("invalid valve status: %s", status)
	}
	resp, err := d.base.sendWTN1604Command(ctx, cmd, 3*time.Second)
	if err != nil {
		return fmt.Errorf("set valve status: %w", err)
	}
	if resp != "A" {
		return fmt.Errorf("set valve status failed: response %q", resp)
	}
	return nil
}

func (d *WTN1604Driver) ReadUnit(ctx context.Context) (string, error) {
	resp, err := d.base.sendWTN1604Command(ctx, "u01101", 3*time.Second)
	if err != nil {
		return "", fmt.Errorf("read unit: %w", err)
	}
	unit, ok := parseWTN1604Unit(resp)
	if !ok {
		return coefficientToUnit(strings.TrimSpace(resp)), nil
	}
	return unit, nil
}

func (d *WTN1604Driver) SetUnit(ctx context.Context, unit string) error {
	coef, ok := unitToCoefficient(unit)
	if !ok {
		return fmt.Errorf("unsupported unit: %s", unit)
	}
	cmd := fmt.Sprintf("v01101 %s", coef)
	resp, err := d.base.sendWTN1604Command(ctx, cmd, 3*time.Second)
	if err != nil {
		return fmt.Errorf("set unit: %w", err)
	}
	if resp != "A" {
		return fmt.Errorf("set unit failed: response %q", resp)
	}
	return nil
}

func (d *WTN1604Driver) CollectData(ctx context.Context, channels []int) ([]float64, error) {
	bitmap := channelsToBitmap(channels)
	cmd := fmt.Sprintf("r%s0", bitmap)
	resp, err := d.base.sendWTN1604Command(ctx, cmd, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("collect data: %w", err)
	}
	if strings.HasPrefix(resp, "N") {
		return nil, fmt.Errorf("device error: %s", resp)
	}
	parts := strings.Fields(resp)
	values := make([]float64, 0, len(parts))
	for _, p := range parts {
		v, err := strconv.ParseFloat(p, 64)
		if err != nil {
			continue
		}
		values = append(values, v)
	}
	for i, j := 0, len(values)-1; i < j; i, j = i+1, j-1 {
		values[i], values[j] = values[j], values[i]
	}
	return values, nil
}

func (d *WTN1604Driver) ReadDeviceInfo(ctx context.Context) (map[string]string, error) {
	info := make(map[string]string)
	resp, err := d.base.sendWTN1604Command(ctx, "A", 3*time.Second)
	if err != nil {
		return nil, fmt.Errorf("device communication test: %w", err)
	}
	info["commTest"] = resp
	if resp, err = d.base.sendWTN1604Command(ctx, "q00", 3*time.Second); err == nil {
		info["model"] = resp
	}
	if resp, err = d.base.sendWTN1604Command(ctx, "q01", 3*time.Second); err == nil {
		info["version"] = resp
	}
	return info, nil
}

func (d *WTN1604Driver) Reset(ctx context.Context) error {
	resp, err := d.base.sendWTN1604Command(ctx, "B", 3*time.Second)
	if err != nil {
		return fmt.Errorf("reset: %w", err)
	}
	if resp != "A" {
		return fmt.Errorf("reset failed: response %q", resp)
	}
	return nil
}

func (d *WTN1604Driver) StartCalibration(ctx context.Context, channels []int, pressurePoints int, avgPoints int) error {
	bitmap := channelsToBitmap(channels)
	cmd := fmt.Sprintf("C 00 %s %d 1 %d", bitmap, pressurePoints, avgPoints)
	resp, err := d.base.sendWTN1604Command(ctx, cmd, 5*time.Second)
	if err != nil {
		return fmt.Errorf("start calibration: %w", err)
	}
	if resp != "A" {
		return fmt.Errorf("start calibration failed: response %q", resp)
	}
	return nil
}

func (d *WTN1604Driver) CollectCalibrationPoint(ctx context.Context, pointIndex int, targetPressure float64) ([]float64, error) {
	cmd := fmt.Sprintf("C 01 %d %.2f", pointIndex, targetPressure)
	resp, err := d.base.sendWTN1604Command(ctx, cmd, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("collect calibration point: %w", err)
	}
	if strings.HasPrefix(resp, "N") {
		return nil, fmt.Errorf("device error: %s", resp)
	}
	parts := strings.Fields(resp)
	values := make([]float64, 0, len(parts))
	for _, p := range parts {
		v, err := strconv.ParseFloat(p, 64)
		if err != nil {
			continue
		}
		values = append(values, v)
	}
	for i, j := 0, len(values)-1; i < j; i, j = i+1, j-1 {
		values[i], values[j] = values[j], values[i]
	}
	return values, nil
}

func (d *WTN1604Driver) PerformFitting(ctx context.Context) error {
	resp, err := d.base.sendWTN1604Command(ctx, "C 02", 10*time.Second)
	if err != nil {
		return fmt.Errorf("perform fitting: %w", err)
	}
	if resp != "A" {
		return fmt.Errorf("perform fitting failed: response %q", resp)
	}
	return nil
}

func (d *WTN1604Driver) EndCalibration(ctx context.Context) error {
	resp, err := d.base.sendWTN1604Command(ctx, "C 03", 5*time.Second)
	if err != nil {
		return fmt.Errorf("end calibration: %w", err)
	}
	if resp != "A" {
		return fmt.Errorf("end calibration failed: response %q", resp)
	}
	return nil
}

func (d *WTN1604Driver) SaveCoefficients(ctx context.Context) error {
	resp, err := d.base.sendWTN1604Command(ctx, "w08", 3*time.Second)
	if err != nil {
		return fmt.Errorf("save zero coefficient: %w", err)
	}
	if resp != "A" {
		return fmt.Errorf("save zero coefficient failed: response %q", resp)
	}
	resp, err = d.base.sendWTN1604Command(ctx, "w09", 3*time.Second)
	if err != nil {
		return fmt.Errorf("save gain coefficient: %w", err)
	}
	if resp != "A" {
		return fmt.Errorf("save gain coefficient failed: response %q", resp)
	}
	return nil
}
