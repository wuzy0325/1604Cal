package session

import (
	"fmt"
	"log"

	"cal1604/internal/device"
	"cal1604/internal/infrastructure/driver"
)

// DriverResolver 提供统一的驱动解析逻辑：优先复用已连接驱动，否则从工厂创建新实例。
type DriverResolver struct {
	DeviceManager  device.DeviceStore
	DriverProvider device.ActiveDriverProvider
	Factory        *driver.Factory
}

// ResolveMeasureDriver 解析计量设备驱动。
func (r *DriverResolver) ResolveMeasureDriver(measureDevID string) (device.MeasureDriver, error) {
	if r.DriverProvider != nil {
		if drv := r.DriverProvider.GetActiveDriver(measureDevID); drv != nil {
			if mDrv, ok := drv.(device.MeasureDriver); ok {
				return mDrv, nil
			}
		}
	}

	measureDev, ok := r.DeviceManager.Get(measureDevID)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrDeviceNotFound, measureDevID)
	}
	return r.Factory.CreateMeasureDriver(measureDev)
}

// ResolvePressureDriver 解析打压设备驱动。
func (r *DriverResolver) ResolvePressureDriver(pressureDevID string) (device.PressureDriver, error) {
	if r.DriverProvider != nil {
		if drv := r.DriverProvider.GetActiveDriver(pressureDevID); drv != nil {
			if pDrv, ok := drv.(device.PressureDriver); ok {
				log.Printf("[session] reuse active pressure driver id=%s type=%T", pressureDevID, pDrv)
				return pDrv, nil
			}
		}
	}

	pressureDev, ok := r.DeviceManager.Get(pressureDevID)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrDeviceNotFound, pressureDevID)
	}
	pDrv, err := r.Factory.CreatePressureDriver(pressureDev)
	if err != nil {
		return nil, err
	}
	log.Printf("[session] create pressure driver id=%s type=%T", pressureDevID, pDrv)
	return pDrv, nil
}
