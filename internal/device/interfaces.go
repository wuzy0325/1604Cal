package device

import (
	"context"

	"cal1604/internal/domain"
)

// PressureDriver 抽象打压设备能力。
type PressureDriver interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	SetTargetPressure(ctx context.Context, target float64) error
	Stop(ctx context.Context) error
	Exhaust(ctx context.Context) error
	ReadCurrentPressure(ctx context.Context) (float64, error)
	ReadUnit(ctx context.Context) (string, error)
	SetUnit(ctx context.Context, unit string) error
	ReadStability(ctx context.Context) (bool, error)
}

// MeasureDriver 抽象计量设备能力。
type MeasureDriver interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	ReadValveStatus(ctx context.Context) (string, error)
	SetValveStatus(ctx context.Context, status string) error
	ReadUnit(ctx context.Context) (string, error)
	SetUnit(ctx context.Context, unit string) error
	CollectData(ctx context.Context, channels []int) ([]float64, error)
	ReadDeviceInfo(ctx context.Context) (map[string]string, error)
	Reset(ctx context.Context) error
}

// DeviceStore 抽象设备配置存储能力。
type DeviceStore interface {
	Upsert(dev domain.Device)
	UpdateStatus(id string, status domain.DeviceStatus) bool
	Delete(id string)
	Get(id string) (domain.Device, bool)
	List() []domain.Device
	CheckUnitConsistency() (bool, []string)
}

// ConnectionDriver 抽象设备连接链路能力。
// 当前阶段只要求覆盖连接与断开，协议命令能力后续分阶段补齐。
type ConnectionDriver interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
}

// ConnectionDriverFactory 按设备模型创建连接驱动。
type ConnectionDriverFactory interface {
	Create(dev domain.Device) (ConnectionDriver, error)
}

// ActiveDriverProvider 返回已连接的设备驱动实例，供校准/测量等服务复用。
// 避免各服务独立创建驱动导致连接丢失。
type ActiveDriverProvider interface {
	GetActiveDriver(id string) ConnectionDriver
}
