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
// MeasureDeviceID 为兼容单设备场景保留（= 首个计量设备 ID），
// MeasureDeviceIDs 为多设备场景的完整设备 ID 集合。
type BindingToken struct {
	MeasureDeviceID  string    `json:"measureDeviceId"`
	MeasureDeviceIDs []string  `json:"measureDeviceIds,omitempty"`
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
// 计量设备支持多台：measureDrivers 以设备 ID 为键存储各驱动，measureDevIDs 保持用户勾选顺序。
type Service struct {
	mu             sync.Mutex
	deviceManager  device.DeviceStore
	factory        device.DriverFactory
	driverProvider device.ActiveDriverProvider
	resolver       *DriverResolver

	measureDrivers map[string]device.MeasureDriver
	measureDevIDs  []string
	pressureDriver device.PressureDriver
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
		measureDrivers: make(map[string]device.MeasureDriver),
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

// BindDevices 绑定计量设备和打压设备到当前会话（单计量设备兼容入口）。
// moduleName 用于标识调用方，防止不同模块间的绑定冲突。
// 同一设备 ID 允许更新（用于 refreshPressure 等临时读取场景）。
// 只有不同模块绑定不同设备时才报错，防止标定和计量相互覆盖对方的设备上下文。
func (s *Service) BindDevices(measureDevID, pressureDevID string, moduleName string) (BindingToken, error) {
	return s.BindMeasureDevices([]string{measureDevID}, pressureDevID, moduleName)
}

// BindMeasureDevices 绑定多台计量设备与一台打压设备到当前会话。
// moduleName 用于标识调用方，防止不同模块间的绑定冲突。
// 同一模块重复提交一致设备集合允许更新；不同模块绑定不同设备时报错。
func (s *Service) BindMeasureDevices(measureDevIDs []string, pressureDevID string, moduleName string) (BindingToken, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(measureDevIDs) == 0 {
		return BindingToken{}, fmt.Errorf("measure device ids must not be empty")
	}

	// 绑定冲突判定：当前会话已被其他模块占用且存在设备集合不一致。
	if s.boundBy != "" && s.boundBy != moduleName && len(s.measureDevIDs) > 0 && !sameIDSet(s.measureDevIDs, measureDevIDs) {
		return BindingToken{}, fmt.Errorf("%w: measure devices %v already bound by %s", ErrDeviceBindingConflict, s.measureDevIDs, s.boundBy)
	}
	if s.pressureDevID != "" && s.pressureDevID != pressureDevID && s.boundBy != "" && s.boundBy != moduleName {
		return BindingToken{}, fmt.Errorf("%w: pressure device %s already bound by %s", ErrDeviceBindingConflict, s.pressureDevID, s.boundBy)
	}

	// 逐台解析计量驱动
	drivers := make(map[string]device.MeasureDriver, len(measureDevIDs))
	for _, id := range measureDevIDs {
		mDrv, err := s.resolver.ResolveMeasureDriver(id)
		if err != nil {
			return BindingToken{}, err
		}
		drivers[id] = mDrv
	}

	var pDrv device.PressureDriver
	if pressureDevID != "" {
		var err error
		pDrv, err = s.resolver.ResolvePressureDriver(pressureDevID)
		if err != nil {
			return BindingToken{}, err
		}
	}

	s.measureDrivers = drivers
	s.measureDevIDs = append([]string(nil), measureDevIDs...)
	s.pressureDevID = pressureDevID
	s.pressureDriver = pDrv
	s.boundBy = moduleName

	s.currentToken = s.buildTokenLocked(moduleName)

	s.publish(events.EventSessionDeviceBound, map[string]any{
		"measureDeviceId":  s.firstMeasureDevIDLocked(),
		"measureDeviceIds": append([]string(nil), s.measureDevIDs...),
		"pressureDeviceId": pressureDevID,
		"boundBy":          moduleName,
	})

	return s.currentToken, nil
}

// BindMeasureDevice 仅绑定单台计量设备驱动（兼容入口）。
func (s *Service) BindMeasureDevice(measureDevID string, moduleName string) (BindingToken, error) {
	return s.BindMeasureDevices([]string{measureDevID}, s.pressureDevID, moduleName)
}

// buildTokenLocked 在锁内构造当前绑定令牌。
func (s *Service) buildTokenLocked(moduleName string) BindingToken {
	ids := append([]string(nil), s.measureDevIDs...)
	return BindingToken{
		MeasureDeviceID:  firstOrEmpty(ids),
		MeasureDeviceIDs: ids,
		PressureDeviceID: s.pressureDevID,
		BoundBy:          moduleName,
		CreatedAt:        time.Now(),
	}
}

// firstMeasureDevIDLocked 在锁内返回首个计量设备 ID。
func (s *Service) firstMeasureDevIDLocked() string {
	return firstOrEmpty(s.measureDevIDs)
}

// firstOrEmpty 返回切片首元素，空切片返回空串。
func firstOrEmpty(ids []string) string {
	if len(ids) == 0 {
		return ""
	}
	return ids[0]
}

// sameIDSet 判断两个设备 ID 集合的元素是否完全一致（忽略顺序）。
func sameIDSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := make(map[string]bool, len(a))
	for _, id := range a {
		seen[id] = true
	}
	for _, id := range b {
		if !seen[id] {
			return false
		}
	}
	return true
}

// validateToken 校验调用方提供的 token 是否匹配当前会话绑定。
// 多设备场景：token 中的每个计量设备 ID 都必须在当前会话集合内。
func (s *Service) validateToken(token BindingToken) error {
	if token.BoundBy == "" || token.CreatedAt.IsZero() {
		return ErrBindingExpired
	}
	if token.BoundBy != s.boundBy {
		return fmt.Errorf("%w: token bound by %q but session bound by %q", ErrBindingExpired, token.BoundBy, s.boundBy)
	}
	ids := token.MeasureDeviceIDs
	if len(ids) == 0 {
		ids = []string{token.MeasureDeviceID}
	}
	if len(ids) == 0 {
		return ErrBindingExpired
	}
	bound := make(map[string]bool, len(s.measureDevIDs))
	for _, id := range s.measureDevIDs {
		bound[id] = true
	}
	for _, id := range ids {
		if !bound[id] {
			return ErrBindingExpired
		}
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

// MeasureDeviceID 返回当前绑定的首个计量设备 ID（兼容单设备场景）。
func (s *Service) MeasureDeviceID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.firstMeasureDevIDLocked()
}

// MeasureDeviceIDs 返回当前绑定的全部计量设备 ID（保持勾选顺序）。
func (s *Service) MeasureDeviceIDs() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.measureDevIDs...)
}

// PressureDeviceID 返回当前绑定的打压设备 ID。
func (s *Service) PressureDeviceID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pressureDevID
}

// MeasureDriver 返回当前绑定的首个计量驱动（供标定服务等内部模块使用）。
func (s *Service) MeasureDriver() device.MeasureDriver {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.measureDevIDs) == 0 {
		return nil
	}
	return s.measureDrivers[s.measureDevIDs[0]]
}

// MeasureDrivers 返回当前绑定的全部计量驱动（设备 ID → 驱动）。
func (s *Service) MeasureDrivers() map[string]device.MeasureDriver {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make(map[string]device.MeasureDriver, len(s.measureDrivers))
	for k, v := range s.measureDrivers {
		result[k] = v
	}
	return result
}

// PressureDriver 返回当前绑定的打压驱动（供标定服务等内部模块使用）。
func (s *Service) PressureDriver() device.PressureDriver {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pressureDriver
}

// measureDriverForLocked 在锁内获取指定设备 ID 的计量驱动。
func (s *Service) measureDriverForLocked(deviceID string) device.MeasureDriver {
	if deviceID == "" {
		return s.measureDrivers[s.firstMeasureDevIDLocked()]
	}
	return s.measureDrivers[deviceID]
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

// ReadMeasureData 从首个计量设备读取实时数据，始终读取全部16通道（兼容入口）。
func (s *Service) ReadMeasureData(ctx context.Context, token BindingToken) ([]float64, error) {
	return s.ReadMeasureDataForDevice(ctx, token, "")
}

// ReadMeasureDataForDevice 从指定计量设备读取实时数据，始终读取全部16通道。
// deviceID 为空时回退到首个计量设备。
func (s *Service) ReadMeasureDataForDevice(ctx context.Context, token BindingToken, deviceID string) ([]float64, error) {
	if err := s.validateToken(token); err != nil {
		return nil, err
	}

	s.mu.Lock()
	drv := s.measureDriverForLocked(deviceID)
	s.mu.Unlock()

	if drv == nil {
		return nil, ErrMeasureDeviceNotSet
	}
	return drv.CollectData(ctx, allChannels)
}

// ReadValveStatus 读取首个计量设备阀门状态（兼容入口）。
func (s *Service) ReadValveStatus(ctx context.Context, token BindingToken) (string, error) {
	return s.ReadValveStatusForDevice(ctx, token, "")
}

// ReadValveStatusForDevice 读取指定计量设备阀门状态。
func (s *Service) ReadValveStatusForDevice(ctx context.Context, token BindingToken, deviceID string) (string, error) {
	if err := s.validateToken(token); err != nil {
		return "", err
	}

	s.mu.Lock()
	drv := s.measureDriverForLocked(deviceID)
	s.mu.Unlock()

	if drv == nil {
		return "", ErrMeasureDeviceNotSet
	}
	return drv.ReadValveStatus(ctx)
}

// SetValveStatus 设置首个计量设备阀门状态（兼容入口）。
func (s *Service) SetValveStatus(ctx context.Context, token BindingToken, status string) error {
	return s.SetValveStatusForDevice(ctx, token, "", status)
}

// SetValveStatusForDevice 设置指定计量设备阀门状态。
func (s *Service) SetValveStatusForDevice(ctx context.Context, token BindingToken, deviceID, status string) error {
	if err := s.validateToken(token); err != nil {
		return err
	}

	s.mu.Lock()
	drv := s.measureDriverForLocked(deviceID)
	s.mu.Unlock()

	if drv == nil {
		return ErrMeasureDeviceNotSet
	}
	return drv.SetValveStatus(ctx, status)
}

// CalibrateZero 对首个计量设备指定通道执行调零校准（兼容入口）。
func (s *Service) CalibrateZero(ctx context.Context, token BindingToken, channels []int) ([]float64, error) {
	return s.CalibrateZeroForDevice(ctx, token, "", channels)
}

// CalibrateZeroForDevice 对指定计量设备通道执行调零校准，并把各通道校零偏移持久化到设备配置，
// 使设备重连后自动加载继续扣除，避免计量数据因零漂漂移。
func (s *Service) CalibrateZeroForDevice(ctx context.Context, token BindingToken, deviceID string, channels []int) ([]float64, error) {
	if err := s.validateToken(token); err != nil {
		return nil, err
	}

	s.mu.Lock()
	if deviceID == "" {
		deviceID = s.firstMeasureDevIDLocked()
	}
	drv := s.measureDriverForLocked(deviceID)
	s.mu.Unlock()

	if drv == nil {
		return nil, ErrMeasureDeviceNotSet
	}

	results, err := drv.CalibrateZero(ctx, channels)
	if err != nil {
		return nil, err
	}
	s.persistTareOffsets(deviceID, channels, results)
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

// CalibrateFullScale 对首个计量设备指定通道执行满量程校准（兼容入口）。
func (s *Service) CalibrateFullScale(ctx context.Context, token BindingToken, channels []int, fullScaleValue float64) ([]float64, error) {
	return s.CalibrateFullScaleForDevice(ctx, token, "", channels, fullScaleValue)
}

// CalibrateFullScaleForDevice 对指定计量设备通道执行满量程校准。
func (s *Service) CalibrateFullScaleForDevice(ctx context.Context, token BindingToken, deviceID string, channels []int, fullScaleValue float64) ([]float64, error) {
	if err := s.validateToken(token); err != nil {
		return nil, err
	}

	s.mu.Lock()
	drv := s.measureDriverForLocked(deviceID)
	s.mu.Unlock()

	if drv == nil {
		return nil, ErrMeasureDeviceNotSet
	}

	return drv.CalibrateFullScale(ctx, channels, fullScaleValue)
}

// ReadMeasureUnit 读取首个计量设备压力单位（兼容入口）。
func (s *Service) ReadMeasureUnit(ctx context.Context, token BindingToken) (string, error) {
	return s.ReadMeasureUnitForDevice(ctx, token, "")
}

// ReadMeasureUnitForDevice 读取指定计量设备压力单位。
// 读取成功后自动将硬件实际单位同步回设备配置存储，确保 CheckUnitConsistency 比较的是硬件真实单位。
func (s *Service) ReadMeasureUnitForDevice(ctx context.Context, token BindingToken, deviceID string) (string, error) {
	if err := s.validateToken(token); err != nil {
		return "", err
	}

	s.mu.Lock()
	if deviceID == "" {
		deviceID = s.firstMeasureDevIDLocked()
	}
	drv := s.measureDriverForLocked(deviceID)
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
	if unit != "" && deviceID != "" {
		if dev, ok := s.deviceManager.Get(deviceID); ok && dev.Unit != unit {
			dev.Unit = unit
			s.deviceManager.Upsert(dev)
			log.Printf("[1604单位读取] 设备 %s 单位已同步: %q → %q", deviceID, dev.Unit, unit)
		}
	}

	return unit, nil
}

// SetMeasureUnit 设置首个计量设备压力单位（兼容入口）。
func (s *Service) SetMeasureUnit(ctx context.Context, token BindingToken, unit string) error {
	return s.SetMeasureUnitForDevice(ctx, token, "", unit)
}

// SetMeasureUnitForDevice 设置指定计量设备压力单位。
func (s *Service) SetMeasureUnitForDevice(ctx context.Context, token BindingToken, deviceID, unit string) error {
	if err := s.validateToken(token); err != nil {
		return err
	}

	s.mu.Lock()
	if deviceID == "" {
		deviceID = s.firstMeasureDevIDLocked()
	}
	drv := s.measureDriverForLocked(deviceID)
	s.mu.Unlock()

	if drv == nil {
		return ErrMeasureDeviceNotSet
	}
	if err := drv.SetUnit(ctx, unit); err != nil {
		return err
	}

	// 同步单位到设备配置存储，保证单位一致性检查读取到最新设定值。
	if deviceID != "" {
		s.deviceManager.UpdateUnit(deviceID, unit)
		log.Printf("[1604单位设置] 计量设备 %s 单位同步为 %q", deviceID, unit)
	}
	return nil
}

// ReadDeviceInfo 读取首个计量设备信息（兼容入口）。
func (s *Service) ReadDeviceInfo(ctx context.Context, token BindingToken) (map[string]string, error) {
	return s.ReadDeviceInfoForDevice(ctx, token, "")
}

// ReadDeviceInfoForDevice 读取指定计量设备信息。
func (s *Service) ReadDeviceInfoForDevice(ctx context.Context, token BindingToken, deviceID string) (map[string]string, error) {
	if err := s.validateToken(token); err != nil {
		return nil, err
	}

	s.mu.Lock()
	drv := s.measureDriverForLocked(deviceID)
	s.mu.Unlock()

	if drv == nil {
		return nil, ErrMeasureDeviceNotSet
	}
	return drv.ReadDeviceInfo(ctx)
}

// ResetDevice 复位首个计量设备（兼容入口）。
func (s *Service) ResetDevice(ctx context.Context, token BindingToken) error {
	return s.ResetDeviceForDevice(ctx, token, "")
}

// ResetDeviceForDevice 复位指定计量设备。
func (s *Service) ResetDeviceForDevice(ctx context.Context, token BindingToken, deviceID string) error {
	if err := s.validateToken(token); err != nil {
		return err
	}

	s.mu.Lock()
	drv := s.measureDriverForLocked(deviceID)
	s.mu.Unlock()

	if drv == nil {
		return ErrMeasureDeviceNotSet
	}
	return drv.Reset(ctx)
}
