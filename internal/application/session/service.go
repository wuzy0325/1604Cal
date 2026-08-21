package session

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"cal1604/internal/device"
	"cal1604/internal/events"
)

// BindingToken 设备绑定租约令牌，标识一次设备绑定的所有权。
// 持有有效 token 的模块才能操作绑定的设备。
type BindingToken struct {
	MeasureDeviceID  string    `json:"measureDeviceId"`
	PressureDeviceID string    `json:"pressureDeviceId,omitempty"`
	BoundBy          string    `json:"boundBy"`
	CreatedAt        time.Time `json:"createdAt"`
}

// allChannels 全部16个通道，用于始终读取全部通道数据。
var allChannels = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

var (
	// ErrMeasureDeviceNotSet 表示计量设备驱动尚未绑定。
	ErrMeasureDeviceNotSet = errors.New("measure device not set")
	// ErrPressureDeviceNotSet 表示打压设备驱动尚未绑定。
	ErrPressureDeviceNotSet = errors.New("pressure device not set")
	// ErrDeviceNotFound 表示设备不存在。
	ErrDeviceNotFound = errors.New("device not found")
	// ErrDeviceBindingConflict 表示设备已被其他模块绑定，不允许覆盖。
	ErrDeviceBindingConflict = errors.New("device binding conflict: device is already bound by another module")
	// ErrBindingExpired 表示绑定令牌已过期或无效。
	ErrBindingExpired = errors.New("binding token expired or invalid")
)

// EventPublisher 广播事件。
type EventPublisher func(eventType string, data any)

// Service 设备会话服务，管理计量设备和打压设备的绑定与实时数据读取。
// 计量和标定模块通过此服务共享设备操作能力。
// 注意：本服务为全局单例，同一时间只能有一组设备绑定。不同模块需要协调使用。
type Service struct {
	mu             sync.Mutex
	deviceManager  device.DeviceStore
	factory        device.DriverFactory
	driverProvider device.ActiveDriverProvider
	resolver       *DriverResolver

	measureDriver  device.MeasureDriver
	pressureDriver device.PressureDriver
	measureDevID   string
	pressureDevID  string
	boundBy        string

	currentToken BindingToken

	publish EventPublisher
}

// NewService 创建设备会话服务。
func NewService(
	deviceManager device.DeviceStore,
	factory device.DriverFactory,
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
// moduleName 用于标识调用方，防止不同模块间的绑定冲突。
// 同一设备 ID 允许更新（用于 refreshPressure 等临时读取场景）。
// 只有不同模块绑定不同设备时才报错，防止标定和计量相互覆盖对方的设备上下文。
func (s *Service) BindDevices(measureDevID, pressureDevID string, moduleName string) (BindingToken, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.measureDevID != "" && s.measureDevID != measureDevID && s.boundBy != "" && s.boundBy != moduleName {
		return BindingToken{}, fmt.Errorf("%w: measure device %s already bound by %s", ErrDeviceBindingConflict, s.measureDevID, s.boundBy)
	}
	if s.pressureDevID != "" && s.pressureDevID != pressureDevID && s.boundBy != "" && s.boundBy != moduleName {
		return BindingToken{}, fmt.Errorf("%w: pressure device %s already bound by %s", ErrDeviceBindingConflict, s.pressureDevID, s.boundBy)
	}

	mDrv, err := s.resolver.ResolveMeasureDriver(measureDevID)
	if err != nil {
		return BindingToken{}, err
	}

	var pDrv device.PressureDriver
	if pressureDevID != "" {
		pDrv, err = s.resolver.ResolvePressureDriver(pressureDevID)
		if err != nil {
			return BindingToken{}, err
		}
	}

	s.measureDevID = measureDevID
	s.pressureDevID = pressureDevID
	s.measureDriver = mDrv
	s.pressureDriver = pDrv
	s.boundBy = moduleName

	s.currentToken = BindingToken{
		MeasureDeviceID:  measureDevID,
		PressureDeviceID: pressureDevID,
		BoundBy:          moduleName,
		CreatedAt:        time.Now(),
	}

	s.publish(events.EventSessionDeviceBound, map[string]any{
		"measureDeviceId":  measureDevID,
		"pressureDeviceId": pressureDevID,
		"boundBy":         moduleName,
	})

	return s.currentToken, nil
}

// BindMeasureDevice 仅绑定计量设备驱动。
func (s *Service) BindMeasureDevice(measureDevID string, moduleName string) (BindingToken, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.measureDevID != "" && s.measureDevID != measureDevID && s.boundBy != "" && s.boundBy != moduleName {
		return BindingToken{}, fmt.Errorf("%w: measure device %s already bound by %s", ErrDeviceBindingConflict, s.measureDevID, s.boundBy)
	}

	mDrv, err := s.resolver.ResolveMeasureDriver(measureDevID)
	if err != nil {
		return BindingToken{}, err
	}

	s.measureDevID = measureDevID
	s.measureDriver = mDrv
	s.boundBy = moduleName

	s.currentToken = BindingToken{
		MeasureDeviceID:  measureDevID,
		PressureDeviceID: s.pressureDevID,
		BoundBy:          moduleName,
		CreatedAt:        time.Now(),
	}

	s.publish(events.EventSessionDeviceBound, map[string]any{
		"measureDeviceId":  measureDevID,
		"pressureDeviceId": s.pressureDevID,
		"boundBy":         moduleName,
	})

	return s.currentToken, nil
}

// validateToken 校验调用方提供的 token 是否匹配当前会话绑定。
func (s *Service) validateToken(token BindingToken) error {
	if token.BoundBy == "" || token.CreatedAt.IsZero() {
		return ErrBindingExpired
	}
	if token.BoundBy != s.boundBy {
		return fmt.Errorf("%w: token bound by %q but session bound by %q", ErrBindingExpired, token.BoundBy, s.boundBy)
	}
	if token.MeasureDeviceID != s.measureDevID {
		return ErrBindingExpired
	}
	return nil
}

// Token 返回当前会话的绑定令牌。
func (s *Service) Token() BindingToken {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.currentToken
}

// CheckUnitConsistency 检查所有已连接设备（计量与打压）的压力单位是否一致。
// 返回是否一致以及单位不一致的设备 ID 列表。
func (s *Service) CheckUnitConsistency() (bool, []string) {
	if s.deviceManager == nil {
		return true, nil
	}
	return s.deviceManager.CheckUnitConsistency()
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
func (s *Service) ReadPressure(ctx context.Context, token BindingToken) (float64, error) {
	if err := s.validateToken(token); err != nil {
		return 0, err
	}

	s.mu.Lock()
	drv := s.pressureDriver
	s.mu.Unlock()

	if drv == nil {
		return 0, ErrPressureDeviceNotSet
	}
	return drv.ReadCurrentPressure(ctx)
}

// ReadStability 读取打压设备稳定状态。
func (s *Service) ReadStability(ctx context.Context, token BindingToken) (bool, error) {
	if err := s.validateToken(token); err != nil {
		return false, err
	}

	s.mu.Lock()
	drv := s.pressureDriver
	s.mu.Unlock()

	if drv == nil {
		return false, ErrPressureDeviceNotSet
	}
	return drv.ReadStability(ctx)
}

// ReadMeasureData 从计量设备读取实时数据，始终读取全部16通道。
func (s *Service) ReadMeasureData(ctx context.Context, token BindingToken) ([]float64, error) {
	if err := s.validateToken(token); err != nil {
		return nil, err
	}

	s.mu.Lock()
	drv := s.measureDriver
	s.mu.Unlock()

	if drv == nil {
		return nil, ErrMeasureDeviceNotSet
	}
	return drv.CollectData(ctx, allChannels)
}

// ReadValveStatus 读取计量设备阀门状态。
func (s *Service) ReadValveStatus(ctx context.Context, token BindingToken) (string, error) {
	if err := s.validateToken(token); err != nil {
		return "", err
	}

	s.mu.Lock()
	drv := s.measureDriver
	s.mu.Unlock()

	if drv == nil {
		return "", ErrMeasureDeviceNotSet
	}
	return drv.ReadValveStatus(ctx)
}

// SetValveStatus 设置计量设备阀门状态。
func (s *Service) SetValveStatus(ctx context.Context, token BindingToken, status string) error {
	if err := s.validateToken(token); err != nil {
		return err
	}

	s.mu.Lock()
	drv := s.measureDriver
	s.mu.Unlock()

	if drv == nil {
		return ErrMeasureDeviceNotSet
	}
	return drv.SetValveStatus(ctx, status)
}

// CalibrateZero 对指定通道执行调零校准，并把各通道校零偏移持久化到设备配置，
// 使设备重连后自动加载继续扣除，避免计量数据因零漂漂移。
func (s *Service) CalibrateZero(ctx context.Context, token BindingToken, channels []int) ([]float64, error) {
	if err := s.validateToken(token); err != nil {
		return nil, err
	}

	s.mu.Lock()
	devID := s.measureDevID
	drv := s.measureDriver
	s.mu.Unlock()

	if drv == nil {
		return nil, ErrMeasureDeviceNotSet
	}

	results, err := drv.CalibrateZero(ctx, channels)
	if err != nil {
		return nil, err
	}
	s.persistTareOffsets(devID, channels, results)
	return results, nil
}

// persistTareOffsets 把校零偏移写回设备配置并持久化（随 devices.json 落盘）。
// channels 与 offsets 一一对应（1-based 通道号 → 校零偏移）。
func (s *Service) persistTareOffsets(devID string, channels []int, offsets []float64) {
	if devID == "" {
		return
	}
	dev, ok := s.deviceManager.Get(devID)
	if !ok {
		return
	}
	idx := make(map[int]int, len(dev.Channels))
	for i := range dev.Channels {
		idx[dev.Channels[i].Index] = i
	}
	for i, ch := range channels {
		pos, found := idx[ch]
		if !found {
			continue
		}
		dev.Channels[pos].TareOffset = offsets[i]
	}
	s.deviceManager.Upsert(dev)
}

// CalibrateFullScale 对指定通道执行满量程校准。
func (s *Service) CalibrateFullScale(ctx context.Context, token BindingToken, channels []int, fullScaleValue float64) ([]float64, error) {
	if err := s.validateToken(token); err != nil {
		return nil, err
	}

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
func (s *Service) ReadMeasureUnit(ctx context.Context, token BindingToken) (string, error) {
	if err := s.validateToken(token); err != nil {
		return "", err
	}

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

	// 同步硬件单位到设备配置存储（仅非空覆盖，避免空响应擦除有效配置）
	if unit != "" && devID != "" {
		if dev, ok := s.deviceManager.Get(devID); ok && dev.Unit != unit {
			dev.Unit = unit
			s.deviceManager.Upsert(dev)
			log.Printf("[1604单位读取] 设备 %s 单位已同步: %q → %q", devID, dev.Unit, unit)
		}
	}

	return unit, nil
}

// SetMeasureUnit 设置计量设备压力单位。
func (s *Service) SetMeasureUnit(ctx context.Context, token BindingToken, unit string) error {
	if err := s.validateToken(token); err != nil {
		return err
	}

	s.mu.Lock()
	drv := s.measureDriver
	devID := s.measureDevID
	s.mu.Unlock()

	if drv == nil {
		return ErrMeasureDeviceNotSet
	}
	if err := drv.SetUnit(ctx, unit); err != nil {
		return err
	}

	// 同步单位到设备配置存储，保证单位一致性检查读取到最新设定值。
	if devID != "" {
		s.deviceManager.UpdateUnit(devID, unit)
		log.Printf("[1604单位设置] 计量设备 %s 单位同步为 %q", devID, unit)
	}
	return nil
}

// ReadDeviceInfo 读取计量设备信息。
func (s *Service) ReadDeviceInfo(ctx context.Context, token BindingToken) (map[string]string, error) {
	if err := s.validateToken(token); err != nil {
		return nil, err
	}

	s.mu.Lock()
	drv := s.measureDriver
	s.mu.Unlock()

	if drv == nil {
		return nil, ErrMeasureDeviceNotSet
	}
	return drv.ReadDeviceInfo(ctx)
}

// ResetDevice 复位计量设备。
func (s *Service) ResetDevice(ctx context.Context, token BindingToken) error {
	if err := s.validateToken(token); err != nil {
		return err
	}

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
