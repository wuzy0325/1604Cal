package driver_test

import (
	"testing"

	"cal1604/internal/domain"
	"cal1604/internal/infrastructure/driver"
)

func TestFactorySupportsMVPModels(t *testing.T) {
	factory := driver.NewFactory()

	testCases := []domain.Device{
		{
			ID:    "m1",
			Type:  domain.DeviceTypeMeasure,
			Model: "WTN1604",
			Host:  "127.0.0.1",
			Port:  9000,
		},
		{
			ID:    "p1",
			Type:  domain.DeviceTypePressure,
			Model: "ConST 811A",
			Host:  "127.0.0.1",
			Port:  7000,
		},
		{
			ID:    "p2",
			Type:  domain.DeviceTypePressure,
			Model: "ConST 820",
			Host:  "127.0.0.1",
			Port:  7001,
		},
	}

	for _, dev := range testCases {
		dev := dev
		t.Run(dev.Model, func(t *testing.T) {
			drv, err := factory.Create(dev)
			if err != nil {
				t.Fatalf("expected model %s to be supported, got error: %v", dev.Model, err)
			}

			if drv == nil {
				t.Fatalf("expected non-nil driver for model %s", dev.Model)
			}
		})
	}
}

func TestFactoryRejectsUnsupportedModel(t *testing.T) {
	factory := driver.NewFactory()

	_, err := factory.Create(domain.Device{
		ID:    "x1",
		Type:  domain.DeviceTypePressure,
		Model: "Unknown Model",
		Host:  "127.0.0.1",
		Port:  7002,
	})
	if err == nil {
		t.Fatal("expected unsupported model error, got nil")
	}
}
