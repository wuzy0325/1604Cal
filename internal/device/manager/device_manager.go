package manager

import (
	"sort"
	"sync"

	"cal1604/internal/domain"
)

// DeviceManager 负责维护设备配置与单位一致性检查。
type DeviceManager struct {
	mu      sync.RWMutex
	devices map[string]domain.Device
}

// NewDeviceManager 创建内存版设备管理器。
func NewDeviceManager() *DeviceManager {
	return &DeviceManager{
		devices: make(map[string]domain.Device),
	}
}

// Upsert 新增或更新设备配置。
func (m *DeviceManager) Upsert(dev domain.Device) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.devices[dev.ID] = dev
}

// UpdateStatus 更新设备连接状态，返回是否更新成功。
func (m *DeviceManager) UpdateStatus(id string, status domain.DeviceStatus) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	dev, ok := m.devices[id]
	if !ok {
		return false
	}

	dev.Status = status
	m.devices[id] = dev
	return true
}

// Delete 删除指定设备。
func (m *DeviceManager) Delete(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.devices, id)
}

// Get 查询指定设备。
func (m *DeviceManager) Get(id string) (domain.Device, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	dev, ok := m.devices[id]
	return dev, ok
}

// List 返回设备快照。
func (m *DeviceManager) List() []domain.Device {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]domain.Device, 0, len(m.devices))
	for _, dev := range m.devices {
		result = append(result, dev)
	}

	return result
}

// CheckUnitConsistency 检查全部设备单位是否一致。
// 返回值依次为：是否一致、冲突设备 ID 列表（升序）。
func (m *DeviceManager) CheckUnitConsistency() (bool, []string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.devices) <= 1 {
		return true, nil
	}

	ids := make([]string, 0, len(m.devices))
	for id := range m.devices {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	baseline := ""
	conflicts := make([]string, 0)

	for _, id := range ids {
		dev := m.devices[id]
		unit := dev.Unit
		if baseline == "" && unit != "" {
			baseline = unit
			continue
		}

		if unit == "" || (baseline != "" && unit != baseline) {
			conflicts = append(conflicts, id)
		}
	}

	if len(conflicts) > 0 {
		sort.Strings(conflicts)
		return false, conflicts
	}

	return true, nil
}
