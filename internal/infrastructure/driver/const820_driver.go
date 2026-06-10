package driver

import (
	"context"
	"fmt"
	"log"
	"time"
)

// ---------------------------------------------------------------------------
// ConST 820 打压设备驱动 (简化SCPI)
// ---------------------------------------------------------------------------

type ConST820Driver struct {
	constBaseDriver
}

func newConST820Driver(host string, port int) *ConST820Driver {
	return &ConST820Driver{
		constBaseDriver: constBaseDriver{base: newTCPConnectionDriver("ConST 820", host, port)},
	}
}

func newConST820DriverWithLocalAddr(host string, port int, localAddr string) *ConST820Driver {
	return &ConST820Driver{
		constBaseDriver: constBaseDriver{base: newTCPConnectionDriverWithLocalAddr("ConST 820", host, port, localAddr)},
	}
}

func (d *ConST820Driver) Connect(ctx context.Context) error {
	return d.constConnect(ctx, "OUTPut:PRESsure:STABle?")
}

func (d *ConST820Driver) SetTargetPressure(ctx context.Context, target float64) error {
	cmd := fmt.Sprintf("SOURce:PRESsure %.4f", target)
	_, err := d.base.sendSCPICommand(ctx, cmd, 3*time.Second)
	if err != nil {
		return fmt.Errorf("set target pressure: %w", err)
	}
	return nil
}

func (d *ConST820Driver) Stop(ctx context.Context) error {
	_, err := d.base.sendSCPICommand(ctx, "OUTPut:PRESsure:MODE VENT", 3*time.Second)
	if err != nil {
		return fmt.Errorf("stop pressure: %w", err)
	}
	return nil
}

func (d *ConST820Driver) Exhaust(ctx context.Context) error {
	_, err := d.base.sendSCPICommand(ctx, "OUTPut:PRESsure:MODE VENT", 3*time.Second)
	if err != nil {
		return fmt.Errorf("exhaust: %w", err)
	}
	return nil
}

func (d *ConST820Driver) ReadCurrentPressure(ctx context.Context) (float64, error) {
	resp, err := d.base.sendSCPICommand(ctx, "MEASure:SCALar:PRESsure1?", 3*time.Second)
	if err != nil {
		return 0, fmt.Errorf("read current pressure: %w", err)
	}
	return parseSCPIPressure(resp)
}

func (d *ConST820Driver) ReadUnit(ctx context.Context) (string, error) {
	resp, err := d.base.sendSCPICommand(ctx, "UNIT:PRESsure?", 3*time.Second)
	if err != nil {
		log.Printf("[820] ReadUnit error: %v", err)
		return "", fmt.Errorf("read unit: %w", err)
	}
	unit := NormalizePressureUnit(parseConST820Unit(resp))
	log.Printf("[820] ReadUnit raw=%q parsed=%q", resp, unit)
	return unit, nil
}

func (d *ConST820Driver) SetUnit(ctx context.Context, unit string) error {
	unitCode, ok := pressureUnitToCode820(unit)
	if !ok {
		return fmt.Errorf("unsupported pressure unit: %s", unit)
	}
	cmd := fmt.Sprintf("UNIT:PRESsure %s", unitCode)
	log.Printf("[820] SetUnit %q → cmd=%q", unit, cmd)
	_, err := d.base.sendSCPICommand(ctx, cmd, 3*time.Second)
	if err != nil {
		return fmt.Errorf("set unit: %w", err)
	}
	return nil
}

func (d *ConST820Driver) ReadStability(ctx context.Context) (bool, error) {
	return d.constReadStability(ctx, "OUTPut:PRESsure:STABle?")
}

func (d *ConST820Driver) IsStable(ctx context.Context) (bool, error) {
	return d.ReadStability(ctx)
}

func (d *ConST820Driver) StartControl(ctx context.Context) error {
	_, err := d.base.sendSCPICommand(ctx, "OUTPut:PRESsure:MODE CONTROL", 3*time.Second)
	if err != nil {
		return fmt.Errorf("start pressure control: %w", err)
	}
	return nil
}


