package session

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"cal1604/internal/device"
	"cal1604/internal/infrastructure/driver"
)

var (
	// ErrMeasureDeviceNotSet 表示计量设备驱动尚未绑定。
	ErrMeasureDeviceNotSet = errors.New("measure device not set")
	// ErrPressureDeviceNotSet 表示打压设备驱动尚未绑定。
	ErrPressureDeviceNotSet = errors.New("pressure device not set")
	// ErrDeviceNotFound 表示设备不存在。
	ErrDeviceNotFound = errors.New("device not found")
)

// EventPublisher 广播事件。
type EventPublisher func(eventType string, data any)

// Service 设备会话服务，管理计量设备和打压设备的绑定与实时数据读取。
// 计量和标定模块通过此服务共享设备操作能力。
type Service struct {
	mu             sync.Mutex
	deviceManager  device.DeviceStore
	factory        *driver.Factory
	driverProvider device.ActiveDriverProvider

	measureDriver  device.MeasureDriver
	pressureDriver device.PressureDriver
	measureDevID   string
	pressureDevID  string

	// channels 用于 ReadMeasureData 时指定通道，可由调用方设置。
	channels []int

	publish EventPublisher
}

// NewService 创建设备会话服务。
func NewService(
	deviceManager device.DeviceStore,
	factory *driver.Factory,
	publisher EventPublisher,
	driverProvider device.ActiveDriverProvider,
) *Service {
	if publisher == nil {
		publisher = func(string, any) {}
	}
	return &Service{
		deviceManager:  deviceManager,
		factory:        factory,
		publish:        publisher,
		driverProvider: driverProvider,
	}
}

// BindDevices 绑定计量设备和打压设备到当前会话。
func (s *Service) BindDevices(measureDevID, pressureDevID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	mDrv, err := s.resolveMeasureDriver(measureDevID)
	if err != nil {
		return err
	}

	var pDrv device.PressureDriver
	if pressureDevID != "" {
		pDrv, err = s.resolvePressureDriver(pressureDevID)
		if err != nil {
			return err
		}
	}

	s.measureDevID = measureDevID
	s.pressureDevID = pressureDevID
	s.measureDriver = mDrv
	s.pressureDriver = pDrv

	s.publish("session.device_bound", map[string]any{
		"measureDeviceId":  measureDevID,
		"pressureDeviceId": pressureDevID,
	})

	return nil
}

// BindMeasureDevice 仅绑定计量设备驱动。
func (s *Service) BindMeasureDevice(measureDevID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	mDrv, err := s.resolveMeasureDriver(measureDevID)
	if err != nil {
		return err
	}

	s.measureDevID = measureDevID
	s.measureDriver = mDrv

	s.publish("session.device_bound", map[string]any{
		"measureDeviceId":  measureDevID,
		"pressureDeviceId": s.pressureDevID,
	})

	return nil
}

// SetChannels 设置默认读取通道列表。
func (s *Service) SetChannels(channels []int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.channels = channels
}

// GetChannels 获取当前通道列表。
func (s *Service) GetChannels() []int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.channels
}

// ReadPressure 读取打压设备当前压力。
func (s *Service) ReadPressure(ctx context.Context) (float64, error) {
	s.mu.Lock()
	drv := s.pressureDriver
	s.mu.Unlock()

	if drv == nil {
		return 0, ErrPressureDeviceNotSet
	}
	return drv.ReadCurrentPressure(ctx)
}

// ReadStability 读取打压设备稳定状态。
func (s *Service) ReadStability(ctx context.Context) (bool, error) {
	s.mu.Lock()
	drv := s.pressureDriver
	s.mu.Unlock()

	if drv == nil {
		return false, ErrPressureDeviceNotSet
	}
	return drv.ReadStability(ctx)
}

// ReadMeasureData 从计量设备读取实时数据。
func (s *Service) ReadMeasureData(ctx context.Context) ([]float64, error) {
	s.mu.Lock()
	drv := s.measureDriver
	channels := s.channels
	s.mu.Unlock()

	if drv == nil {
		return nil, ErrMeasureDeviceNotSet
	}
	if len(channels) == 0 {
		channels = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	}
	return drv.CollectData(ctx, channels)
}

// ReadValveStatus 读取计量设备阀门状态。
func (s *Service) ReadValveStatus(ctx context.Context) (string, error) {
	s.mu.Lock()
	drv := s.measureDriver
	s.mu.Unlock()

	if drv == nil {
		return "", ErrMeasureDeviceNotSet
	}
	return drv.ReadValveStatus(ctx)
}

// SetValveStatus 设置计量设备阀门状态。
func (s *Service) SetValveStatus(ctx context.Context, status string) error {
	s.mu.Lock()
	drv := s.measureDriver
	s.mu.Unlock()

	if drv == nil {
		return ErrMeasureDeviceNotSet
	}
	return drv.SetValveStatus(ctx, status)
}

// ReadMeasureUnit 读取计量设备压力单位。
func (s *Service) ReadMeasureUnit(ctx context.Context) (string, error) {
	s.mu.Lock()
	drv := s.measureDriver
	s.mu.Unlock()

	if drv == nil {
		return "", ErrMeasureDeviceNotSet
	}
	return drv.ReadUnit(ctx)
}

// SetMeasureUnit 设置计量设备压力单位。
func (s *Service) SetMeasureUnit(ctx context.Context, unit string) error {
	s.mu.Lock()
	drv := s.measureDriver
	s.mu.Unlock()

	if drv == nil {
		return ErrMeasureDeviceNotSet
	}
	return drv.SetUnit(ctx, unit)
}

// ReadDeviceInfo 读取计量设备信息。
func (s *Service) ReadDeviceInfo(ctx context.Context) (map[string]string, error) {
	s.mu.Lock()
	drv := s.measureDriver
	s.mu.Unlock()

	if drv == nil {
		return nil, ErrMeasureDeviceNotSet
	}
	return drv.ReadDeviceInfo(ctx)
}

// ResetDevice 复位计量设备。
func (s *Service) ResetDevice(ctx context.Context) error {
	s.mu.Lock()
	drv := s.measureDriver
	s.mu.Unlock()

	if drv == nil {
		return ErrMeasureDeviceNotSet
	}
	return drv.Reset(ctx)
}

// MeasureDriver 返回当前绑定的计量驱动（供标定服务等内部模块使用）。
func (s *Service) MeasureDriver() device.MeasureDriver {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.measureDriver
}

// PressureDriver 返回当前绑定的打压驱动（供标定服务等内部模块使用）。
func (s *Service) PressureDriver() device.PressureDriver {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pressureDriver
}

// resolveMeasureDriver 优先复用已连接的计量驱动。
func (s *Service) resolveMeasureDriver(measureDevID string) (device.MeasureDriver, error) {
	if s.driverProvider != nil {
		if drv := s.driverProvider.GetActiveDriver(measureDevID); drv != nil {
			if mDrv, ok := drv.(device.MeasureDriver); ok {
				return mDrv, nil
			}
		}
	}

	measureDev, ok := s.deviceManager.Get(measureDevID)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrDeviceNotFound, measureDevID)
	}
	return s.factory.CreateMeasureDriver(measureDev)
}

// resolvePressureDriver 优先复用已连接的打压驱动。
func (s *Service) resolvePressureDriver(pressureDevID string) (device.PressureDriver, error) {
	if s.driverProvider != nil {
		if drv := s.driverProvider.GetActiveDriver(pressureDevID); drv != nil {
			if pDrv, ok := drv.(device.PressureDriver); ok {
				return pDrv, nil
			}
		}
	}

	pressureDev, ok := s.deviceManager.Get(pressureDevID)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrDeviceNotFound, pressureDevID)
	}
	return s.factory.CreatePressureDriver(pressureDev)
}
