package manager_test

import (
	"testing"

	"cal1604/internal/device/manager"
	"cal1604/internal/domain"
)

func TestCheckUnitConsistency(t *testing.T) {
	mgr := manager.NewDeviceManager()

	mgr.Upsert(domain.Device{ID: "m1", Type: domain.DeviceTypeMeasure, Unit: "kPa"})
	mgr.Upsert(domain.Device{ID: "p1", Type: domain.DeviceTypePressure, Unit: "kPa"})

	ok, conflictIDs := mgr.CheckUnitConsistency()
	if !ok {
		t.Fatalf("expected consistent units, got conflicts: %v", conflictIDs)
	}
}

func TestCheckUnitConsistencyWithConflict(t *testing.T) {
	mgr := manager.NewDeviceManager()

	mgr.Upsert(domain.Device{ID: "m1", Type: domain.DeviceTypeMeasure, Unit: "kPa"})
	mgr.Upsert(domain.Device{ID: "p1", Type: domain.DeviceTypePressure, Unit: "psi"})

	ok, conflictIDs := mgr.CheckUnitConsistency()
	if ok {
		t.Fatal("expected unit consistency check to fail")
	}

	if len(conflictIDs) == 0 {
		t.Fatal("expected conflict device ids")
	}
}
