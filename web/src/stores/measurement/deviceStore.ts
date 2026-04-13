import { defineStore } from 'pinia'
import { ref } from 'vue'
import {
  fetchDevices,
  connectDevice,
  disconnectDevice,
  upsertDevice,
  type DeviceDTO
} from '@/services/apiClient'
import { ElMessage } from 'element-plus'

// 前端设备模型
export interface PressureDevice {
  id: string
  name: string
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
  status: 'connected' | 'disconnected' | 'connecting' | 'error'
}

// DTO转换函数
function dtoToPressureDevice(dto: DeviceDTO): PressureDevice {
  return {
    id: dto.id,
    name: dto.name,
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
    model: device.name,
    host: device.ip,
    port: device.port,
    unit: device.unit,
    status: device.status === 'connected' ? 'connected' : 'disconnected'
  }
}

function measureDeviceToDto(device: MeasureDevice): DeviceDTO {
  return {
    id: device.id,
    name: device.name,
    type: 'measure',
    model: device.model,
    host: '', // 计量设备可能没有host
    port: 0,
    unit: 'kPa',
    status: device.status === 'connected' ? 'connected' : 'disconnected'
  }
}

export const useMeasurementDeviceStore = defineStore('measurementDevices', () => {
  // State
  const pressureDevices = ref<PressureDevice[]>([])
  const measureDevices = ref<MeasureDevice[]>([])
  const loading = ref(false)

  // 从后端加载设备列表
  const loadDevices = async () => {
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
      ElMessage.error('加载设备列表失败')
    } finally {
      loading.value = false
    }
  }

  // Actions
  const connectPressureDevice = async (id: string) => {
    const device = pressureDevices.value.find(d => d.id === id)
    if (!device) return

    try {
      device.status = 'connecting'
      const updatedDto = await connectDevice(id)
      const updated = dtoToPressureDevice(updatedDto)
      Object.assign(device, updated)
      ElMessage.success(`设备 ${device.name} 连接成功`)
    } catch (error) {
      device.status = 'error'
      console.error('连接设备失败:', error)
      ElMessage.error(`连接设备 ${device.name} 失败`)
    }
  }

  const disconnectPressureDevice = async (id: string) => {
    const device = pressureDevices.value.find(d => d.id === id)
    if (!device) return

    try {
      const updatedDto = await disconnectDevice(id)
      const updated = dtoToPressureDevice(updatedDto)
      Object.assign(device, updated)
      ElMessage.success(`设备 ${device.name} 已断开`)
    } catch (error) {
      console.error('断开设备失败:', error)
      ElMessage.error(`断开设备 ${device.name} 失败`)
    }
  }

  const connectMeasureDevice = async (id: string) => {
    const device = measureDevices.value.find(d => d.id === id)
    if (!device) return

    try {
      device.status = 'connecting'
      const updatedDto = await connectDevice(id)
      const updated = dtoToMeasureDevice(updatedDto)
      Object.assign(device, updated)
      ElMessage.success(`设备 ${device.name} 连接成功`)
    } catch (error) {
      device.status = 'error'
      console.error('连接设备失败:', error)
      ElMessage.error(`连接设备 ${device.name} 失败`)
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
    updateDeviceStatus
  }
})
