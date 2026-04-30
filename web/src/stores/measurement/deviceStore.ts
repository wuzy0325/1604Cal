import { defineStore } from 'pinia'
import { ref } from 'vue'
import {
  fetchDevices,
  connectDevice,
  disconnectDevice,
  upsertDevice
} from "@/api/device"
import {
  bindDevices as bindSessionDevices,
  readPressure as readSessionPressure
} from "@/api/session"
import {
  multipressRegister,
  multipressUnregister,
  multipressListDevices
} from "@/api/multipress"
import type { DeviceDTO } from "@/types/device"
import { ElMessage } from 'element-plus'

// 前端设备模型
export interface PressureDevice {
  id: string
  name: string
  model: string
  ip: string
  port: number
  status: 'connected' | 'disconnected' | 'connecting' | 'error'
  currentPressure?: number
  unit: string
}

export interface MeasureDevice {
  id: string
  name: string
  model: string
  channels: number
  ip?: string
  port?: number
  status: 'connected' | 'disconnected' | 'connecting' | 'error'
}

// DTO转换函数
function dtoToPressureDevice(dto: DeviceDTO): PressureDevice {
  return {
    id: dto.id,
    name: dto.name,
    model: dto.model,
    ip: dto.host,
    port: dto.port,
    status: dto.status === 'connected' ? 'connected' :
            dto.status === 'connecting' ? 'connecting' :
            dto.status === 'error' ? 'error' : 'disconnected',
    currentPressure: dto.status === 'connected' ? 0 : undefined,
    unit: dto.unit || 'kPa'
  }
}

function dtoToMeasureDevice(dto: DeviceDTO): MeasureDevice {
  return {
    id: dto.id,
    name: dto.name,
    model: dto.model,
    channels: 16, // 默认为16通道
    status: dto.status === 'connected' ? 'connected' : 
            dto.status === 'connecting' ? 'connecting' :
            dto.status === 'error' ? 'error' : 'disconnected'
  }
}

function pressureDeviceToDto(device: PressureDevice): DeviceDTO {
  return {
    id: device.id,
    name: device.name,
    type: 'pressure',
    model: device.model,
    host: device.ip,
    port: device.port,
    unit: device.unit,
    status: device.status === 'connected' ? 'connected' :
            device.status === 'connecting' ? 'connecting' :
            device.status === 'error' ? 'error' : 'disconnected'
  }
}

function measureDeviceToDto(device: MeasureDevice): DeviceDTO {
  return {
    id: device.id,
    name: device.name,
    type: 'measure',
    model: device.model,
    host: device.ip || '192.168.1.100',
    port: device.port || 9000,
    unit: 'kPa',
    status: device.status === 'connected' ? 'connected' :
            device.status === 'connecting' ? 'connecting' :
            device.status === 'error' ? 'error' : 'disconnected'
  }
}

export const useMeasurementDeviceStore = defineStore('measurementDevices', () => {
  // State
  const pressureDevices = ref<PressureDevice[]>([])
  const measureDevices = ref<MeasureDevice[]>([])
  const loading = ref(false)

  // 从后端加载设备列表
  const loadDevices = async (silent = false) => {
    try {
      loading.value = true
      const devices = await fetchDevices()
      pressureDevices.value = devices
        .filter(d => d.type === 'pressure')
        .map(dtoToPressureDevice)
      measureDevices.value = devices
        .filter(d => d.type === 'measure')
        .map(dtoToMeasureDevice)
    } catch (error) {
      console.error('加载设备列表失败:', error)
      if (!silent) {
        ElMessage.error('加载设备列表失败')
      }
    } finally {
      loading.value = false
    }
  }

  // Actions
  const connectPressureDevice = async (id: string): Promise<boolean> => {
    const device = pressureDevices.value.find(d => d.id === id)
    if (!device) return false

    try {
      device.status = 'connecting'
      await multipressRegister(id)

      device.status = 'connected'
      if (typeof device.currentPressure !== 'number') {
        device.currentPressure = 0
      }

      // 注册后从 multipress 服务拉取实际读取到的单位与压力，覆盖配置中的默认值
      try {
        const states = await multipressListDevices()
        const state = states.find(s => s.deviceId === id)
        if (state) {
          if (state.unit) {
            device.unit = state.unit
          }
          device.currentPressure = state.currentPressure
        }
      } catch {
        // 静默失败，使用设备配置中的单位
      }

      ElMessage.success(`设备 ${device.name} 连接成功`)

      // 连接成功后尝试读取初始压力值。
      await refreshPressureForDevice(id)
      return true
    } catch (error) {
      device.status = 'error'
      console.error('连接设备失败:', error)
      ElMessage.error(`连接设备 ${device.name} 失败`)
      return false
    }
  }

  // 刷新打压设备的实时压力值
  const refreshPressureForDevice = async (pressureId: string) => {
    try {
      // 优先选择已连接的计量设备，回退到列表中的第一个设备。
      const anyMeasure = measureDevices.value.find(d => d.status === 'connected') ?? measureDevices.value[0]
      const measureId = anyMeasure?.id || '__none__'
      if (measureId === '__none__') return

      await bindSessionDevices(measureId, pressureId)
      const pressure = await readSessionPressure()
      const device = pressureDevices.value.find(d => d.id === pressureId)
      if (device) {
        device.currentPressure = pressure
      }
    } catch (err) {
      console.warn('读取初始压力失败:', err)
    }
  }

  // 刷新所有已连接打压设备的压力值
  const refreshAllConnectedPressures = async () => {
    for (const device of pressureDevices.value) {
      if (device.status === 'connected') {
        try {
          const anyMeasure = measureDevices.value.find(d => d.status === 'connected') ?? measureDevices.value[0]
          if (!anyMeasure) continue
          await bindSessionDevices(anyMeasure.id, device.id)
          const pressure = await readSessionPressure()
          device.currentPressure = pressure
        } catch {
          // 静默失败，不影响其他设备
        }
      }
    }
  }

  const disconnectPressureDevice = async (id: string) => {
    const device = pressureDevices.value.find(d => d.id === id)
    if (!device) return

    try {
      await multipressUnregister(id)
      device.status = 'disconnected'
      device.currentPressure = undefined
      ElMessage.success(`设备 ${device.name} 已断开`)
    } catch (error) {
      console.error('断开设备失败:', error)
      ElMessage.error(`断开设备 ${device.name} 失败`)
    }
  }

  const connectMeasureDevice = async (id: string): Promise<boolean> => {
    const device = measureDevices.value.find(d => d.id === id)
    if (!device) return false

    try {
      device.status = 'connecting'
      const updatedDto = await connectDevice(id)
      const updated = dtoToMeasureDevice(updatedDto)
      Object.assign(device, updated)

      if (updated.status !== 'connected') {
        const reason = updatedDto.lastErrorReason || '未知原因'
        ElMessage.error(`连接设备 ${device.name} 失败：${reason}`)
        return false
      }

      ElMessage.success(`设备 ${device.name} 连接成功`)
      return true
    } catch (error) {
      device.status = 'error'
      console.error('连接设备失败:', error)
      ElMessage.error(`连接设备 ${device.name} 失败`)
      return false
    }
  }

  const disconnectMeasureDevice = async (id: string) => {
    const device = measureDevices.value.find(d => d.id === id)
    if (!device) return

    try {
      const updatedDto = await disconnectDevice(id)
      const updated = dtoToMeasureDevice(updatedDto)
      Object.assign(device, updated)
      ElMessage.success(`设备 ${device.name} 已断开`)
    } catch (error) {
      console.error('断开设备失败:', error)
      ElMessage.error(`断开设备 ${device.name} 失败`)
    }
  }

  // 添加新设备
  const addPressureDevice = async (device: Omit<PressureDevice, 'id' | 'status'>) => {
    try {
      const dto = await upsertDevice({
        ...pressureDeviceToDto({ ...device, id: '', status: 'disconnected' }),
        id: crypto.randomUUID()
      })
      pressureDevices.value.push(dtoToPressureDevice(dto))
      ElMessage.success('设备添加成功')
    } catch (error) {
      console.error('添加设备失败:', error)
      ElMessage.error('添加设备失败')
      throw error
    }
  }

  const addMeasureDevice = async (device: Omit<MeasureDevice, 'id' | 'status'>) => {
    try {
      const dto = await upsertDevice({
        ...measureDeviceToDto({ ...device, id: '', status: 'disconnected' }),
        id: crypto.randomUUID()
      })
      measureDevices.value.push(dtoToMeasureDevice(dto))
      ElMessage.success('设备添加成功')
    } catch (error) {
      console.error('添加设备失败:', error)
      ElMessage.error('添加设备失败')
      throw error
    }
  }

  // 更新设备压力值（用于SSE更新）
  const updateDevicePressure = (id: string, pressure: number) => {
    const device = pressureDevices.value.find(d => d.id === id)
    if (device) {
      device.currentPressure = pressure
    }
  }

  // 更新设备状态（用于SSE更新）
  const updateDeviceStatus = (id: string, status: PressureDevice['status']) => {
    const pressureDevice = pressureDevices.value.find(d => d.id === id)
    const measureDevice = measureDevices.value.find(d => d.id === id)
    
    if (pressureDevice) {
      pressureDevice.status = status
      if (status === 'connected') {
        pressureDevice.currentPressure = 0
      } else if (status === 'disconnected') {
        pressureDevice.currentPressure = undefined
      }
    }
    
    if (measureDevice) {
      measureDevice.status = status
    }
  }

  return {
    pressureDevices,
    measureDevices,
    loading,
    loadDevices,
    connectPressureDevice,
    disconnectPressureDevice,
    connectMeasureDevice,
    disconnectMeasureDevice,
    addPressureDevice,
    addMeasureDevice,
    updateDevicePressure,
    updateDeviceStatus,
    refreshPressureForDevice,
    refreshAllConnectedPressures
  }
})
