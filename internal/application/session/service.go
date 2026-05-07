package session

import (
	"context"
	"errors"
	"log"
	"sync"

	"cal1604/internal/device"
	"cal1604/internal/events"
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
	resolver       *DriverResolver

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
	s := &Service{
		deviceManager:  deviceManager,
		factory:        factory,
		publish:        publisher,
		driverProvider: driverProvider,
	}
	s.resolver = &DriverResolver{
		DeviceManager:  deviceManager,
		DriverProvider: driverProvider,
		Factory:        factory,
	}
	return s
}

// Resolver 返回内部驱动解析器，供 calibration 等模块复用驱动解析逻辑。
func (s *Service) Resolver() *DriverResolver {
	return s.resolver
}

// BindDevices 绑定计量设备和打压设备到当前会话。
func (s *Service) BindDevices(measureDevID, pressureDevID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	mDrv, err := s.resolver.ResolveMeasureDriver(measureDevID)
	if err != nil {
		return err
	}

	var pDrv device.PressureDriver
	if pressureDevID != "" {
		pDrv, err = s.resolver.ResolvePressureDriver(pressureDevID)
		if err != nil {
			return err
		}
	}

	s.measureDevID = measureDevID
	s.pressureDevID = pressureDevID
	s.measureDriver = mDrv
	s.pressureDriver = pDrv

	s.publish(events.EventSessionDeviceBound, map[string]any{
		"measureDeviceId":  measureDevID,
		"pressureDeviceId": pressureDevID,
	})

	return nil
}

// BindMeasureDevice 仅绑定计量设备驱动。
func (s *Service) BindMeasureDevice(measureDevID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	mDrv, err := s.resolver.ResolveMeasureDriver(measureDevID)
	if err != nil {
		return err
	}

	s.measureDevID = measureDevID
	s.measureDriver = mDrv

	s.publish(events.EventSessionDeviceBound, map[string]any{
		"measureDeviceId":  measureDevID,
		"pressureDeviceId": s.pressureDevID,
	})

	return nil
}

// SetChannels 设置默认读取通道列表。
func (s *Service) SetChannels(channels []int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.channels = append([]int(nil), channels...)
}

// GetChannels 获取当前通道列表。
func (s *Service) GetChannels() []int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]int(nil), s.channels...)
}

// MeasureDeviceID 返回当前绑定的计量设备 ID。
func (s *Service) MeasureDeviceID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.measureDevID
}

// PressureDeviceID 返回当前绑定的打压设备 ID。
func (s *Service) PressureDeviceID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pressureDevID
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

// CalibrateZero 对指定通道执行调零校准。
func (s *Service) CalibrateZero(ctx context.Context, channels []int) ([]float64, error) {
	s.mu.Lock()
	drv := s.measureDriver
	s.mu.Unlock()

	if drv == nil {
		return nil, ErrMeasureDeviceNotSet
	}

	return drv.CalibrateZero(ctx, channels)
}

// CalibrateFullScale 对指定通道执行满量程校准。
func (s *Service) CalibrateFullScale(ctx context.Context, channels []int, fullScaleValue float64) ([]float64, error) {
	s.mu.Lock()
	drv := s.measureDriver
	s.mu.Unlock()

	if drv == nil {
		return nil, ErrMeasureDeviceNotSet
	}

	return drv.CalibrateFullScale(ctx, channels, fullScaleValue)
}

// ReadMeasureUnit 读取计量设备压力单位。
// 读取成功后自动将硬件实际单位同步回设备配置存储，确保 CheckUnitConsistency 比较的是硬件真实单位。
func (s *Service) ReadMeasureUnit(ctx context.Context) (string, error) {
	s.mu.Lock()
	drv := s.measureDriver
	devID := s.measureDevID
	s.mu.Unlock()

	if drv == nil {
		return "", ErrMeasureDeviceNotSet
	}
	unit, err := drv.ReadUnit(ctx)
	if err != nil {
		log.Printf("[1604单位读取] 从硬件读取失败: %v", err)
		return "", err
	}
	log.Printf("[1604单位读取] 从硬件读取到单位: %s", unit)

	// 同步硬件单位到设备配置存储
	if devID != "" {
		if dev, ok := s.deviceManager.Get(devID); ok && dev.Unit != unit {
			dev.Unit = unit
			s.deviceManager.Upsert(dev)
			log.Printf("[1604单位读取] 设备 %s 单位已同步: %q → %q", devID, dev.Unit, unit)
		}
	}

	return unit, nil
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
