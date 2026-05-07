package domain

import (
	"errors"
	"net"
	"time"
)

// DeviceType 表示设备类型。
type DeviceType string

const (
	// DeviceTypePressure 表示打压设备。
	DeviceTypePressure DeviceType = "pressure"
	// DeviceTypeMeasure 表示计量设备。
	DeviceTypeMeasure DeviceType = "measure"
)

var ErrInvalidDevice = errors.New("invalid device parameters")

// IsValid 报告设备类型是否合法。
func (t DeviceType) IsValid() bool {
	return t == DeviceTypePressure || t == DeviceTypeMeasure
}

// DeviceStatus 表示设备连接状态。
type DeviceStatus string

const (
	DeviceStatusDisconnected DeviceStatus = "disconnected"
	DeviceStatusConnecting   DeviceStatus = "connecting"
	DeviceStatusConnected    DeviceStatus = "connected"
	DeviceStatusError        DeviceStatus = "error"
)

// IsValid 报告设备状态是否合法。
func (s DeviceStatus) IsValid() bool {
	return s == DeviceStatusDisconnected || s == DeviceStatusConnecting ||
		s == DeviceStatusConnected || s == DeviceStatusError
}

// Validate 校验设备实体字段是否满足基本约束。
func (d Device) Validate() error {
	if d.ID == "" || !d.Type.IsValid() {
		return ErrInvalidDevice
	}
	if net.ParseIP(d.Host) == nil || d.Port < 1 || d.Port > 65535 {
		return ErrInvalidDevice
	}
	return nil
}

// ResolveStatus 根据请求状态和已有设备记录确定最终状态。
// 若请求未指定状态，则继承已有设备的状态，或默认为 Disconnected。
func ResolveStatus(requested DeviceStatus, existing Device, existed bool) DeviceStatus {
	if requested != "" {
		return requested
	}
	if existed && existing.Status != "" {
		return existing.Status
	}
	return DeviceStatusDisconnected
}

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
