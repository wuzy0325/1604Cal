package driver

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"
)

// ---------------------------------------------------------------------------
// SPC4000 打压设备驱动 (Mensor兼容命令集)
// ---------------------------------------------------------------------------

type SPC4000Driver struct {
	base *tcpConnectionDriver
}

func newSPC4000Driver(host string, port int) *SPC4000Driver {
	return &SPC4000Driver{base: newTCPConnectionDriver("SPC4000", host, port)}
}

func (d *SPC4000Driver) Connect(ctx context.Context) error    { return d.base.Connect(ctx) }
func (d *SPC4000Driver) Disconnect(ctx context.Context) error { return d.base.Disconnect(ctx) }

func (d *SPC4000Driver) SetTargetPressure(ctx context.Context, target float64) error {
	var cmd string
	if target >= 0 {
		cmd = fmt.Sprintf("GP %.4f", target)
	} else {
		cmd = fmt.Sprintf("GN %.4f", math.Abs(target))
	}
	_, err := d.base.sendSCPICommand(ctx, cmd, 3*time.Second)
	if err != nil {
		return fmt.Errorf("set target pressure: %w", err)
	}
	return nil
}

func (d *SPC4000Driver) Stop(ctx context.Context) error {
	_, err := d.base.sendSCPICommand(ctx, "Measure", 3*time.Second)
	if err != nil {
		return fmt.Errorf("stop pressure: %w", err)
	}
	return nil
}

func (d *SPC4000Driver) Exhaust(ctx context.Context) error {
	_, err := d.base.sendSCPICommand(ctx, "Vent", 3*time.Second)
	if err != nil {
		return fmt.Errorf("exhaust: %w", err)
	}
	return nil
}

func (d *SPC4000Driver) ReadCurrentPressure(ctx context.Context) (float64, error) {
	resp, err := d.base.sendSCPICommand(ctx, "RP", 3*time.Second)
	if err != nil {
		return 0, fmt.Errorf("read current pressure: %w", err)
	}
	return parseSCPIPressure(resp)
}

func (d *SPC4000Driver) ReadUnit(ctx context.Context) (string, error) {
	resp, err := d.base.sendSCPICommand(ctx, "Units?", 3*time.Second)
	if err != nil {
		log.Printf("[SPC4000] ReadUnit error: %v", err)
		return "", fmt.Errorf("read unit: %w", err)
	}
	unit := NormalizePressureUnit(parseSPC4000UnitCode(resp))
	log.Printf("[SPC4000] ReadUnit raw=%q parsed=%q", resp, unit)
	return unit, nil
}

func (d *SPC4000Driver) SetUnit(ctx context.Context, unit string) error {
	code, ok := pressureUnitToCodeSPC4000(unit)
	if !ok {
		return fmt.Errorf("unsupported pressure unit: %s", unit)
	}
	cmd := fmt.Sprintf("Units %s", code)
	log.Printf("[SPC4000] SetUnit %q → cmd=%q", unit, cmd)
	_, err := d.base.sendSCPICommand(ctx, cmd, 3*time.Second)
	if err != nil {
		return fmt.Errorf("set unit: %w", err)
	}
	return nil
}

func (d *SPC4000Driver) ReadStability(ctx context.Context) (bool, error) { return true, nil }

func (d *SPC4000Driver) StartControl(ctx context.Context) error { return nil }

func (d *SPC4000Driver) ReadTargetRange(ctx context.Context) (min, max float64, err error) {
	return 0, 0, fmt.Errorf("SPC4000 does not support target range query")
}
