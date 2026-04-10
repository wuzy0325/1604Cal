package driver

import (
	"fmt"
	"strings"

	"cal1604/internal/device"
	"cal1604/internal/domain"
)

// Factory 按设备型号创建对应连接驱动（Adapter + Factory）。
type Factory struct{}

// NewFactory 创建连接驱动工厂。
func NewFactory() *Factory {
	return &Factory{}
}

// Create 根据设备模型返回连接驱动实例。
func (f *Factory) Create(dev domain.Device) (device.ConnectionDriver, error) {
	model := normalizeModel(dev.Model)
	switch model {
	case "WTN1604":
		return newWTN1604Driver(dev.Host, dev.Port), nil
	case "CONST811A", "811A":
		return newConST811ADriver(dev.Host, dev.Port), nil
	case "CONST820", "820":
		return newConST820Driver(dev.Host, dev.Port), nil
	default:
		return nil, fmt.Errorf("unsupported device model: %s", dev.Model)
	}
}

func normalizeModel(raw string) string {
	trimmed := strings.TrimSpace(raw)
	upper := strings.ToUpper(trimmed)
	return strings.ReplaceAll(upper, " ", "")
}
