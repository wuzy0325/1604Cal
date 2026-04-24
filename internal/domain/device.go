package domain

import "time"

// DeviceType 表示设备类型。
type DeviceType string

const (
	// DeviceTypePressure 表示打压设备。
	DeviceTypePressure DeviceType = "pressure"
	// DeviceTypeMeasure 表示计量设备。
	DeviceTypeMeasure DeviceType = "measure"
)

// DeviceStatus 表示设备连接状态。
type DeviceStatus string

const (
	DeviceStatusDisconnected DeviceStatus = "disconnected"
	DeviceStatusConnecting   DeviceStatus = "connecting"
	DeviceStatusConnected    DeviceStatus = "connected"
	DeviceStatusError        DeviceStatus = "error"
)

// Device 表示系统维护的设备配置实体。
type Device struct {
	ID     string       `json:"id"`
	Name   string       `json:"name"`
	Type   DeviceType   `json:"type"`
	Model  string       `json:"model"`
	Host   string       `json:"host"`
	Port   int          `json:"port"`
	Unit   string       `json:"unit"`
	Status DeviceStatus `json:"status"`

	// IsSimulated 为 true 时使用模拟驱动，不连接真实设备。
	IsSimulated bool `json:"isSimulated"`

	// LastErrorReason 记录最近一次连接/断连失败原因。
	LastErrorReason string `json:"lastErrorReason,omitempty"`
	// LastErrorAt 记录最近一次连接/断连失败时间。
	LastErrorAt *time.Time `json:"lastErrorAt,omitempty"`
}
