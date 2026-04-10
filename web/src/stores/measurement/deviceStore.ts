import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface PressureDevice {
  id: string
  name: string
  ip: string
  port: number
  status: 'connected' | 'disconnected'
  currentPressure?: number
  unit: string
}

export interface MeasureDevice {
  id: string
  name: string
  model: string
  channels: number
  status: 'connected' | 'disconnected'
}

export const useMeasurementDeviceStore = defineStore('measurementDevices', () => {
  // State
  const pressureDevices = ref<PressureDevice[]>([
    { id: '1', name: '打压设备-1', ip: '192.168.1.100', port: 502, status: 'disconnected', unit: 'kPa' }
  ])
  
  const measureDevices = ref<MeasureDevice[]>([
    { id: '1', name: '计量设备-1', model: '1604', channels: 16, status: 'disconnected' }
  ])
  
  // Actions
  const connectPressureDevice = (id: string) => {
    const device = pressureDevices.value.find(d => d.id === id)
    if (device) {
      device.status = 'connected'
      device.currentPressure = 0
    }
  }
  
  const disconnectPressureDevice = (id: string) => {
    const device = pressureDevices.value.find(d => d.id === id)
    if (device) {
      device.status = 'disconnected'
      device.currentPressure = undefined
    }
  }
  
  const connectMeasureDevice = (id: string) => {
    const device = measureDevices.value.find(d => d.id === id)
    if (device) {
      device.status = 'connected'
    }
  }
  
  const disconnectMeasureDevice = (id: string) => {
    const device = measureDevices.value.find(d => d.id === id)
    if (device) {
      device.status = 'disconnected'
    }
  }
  
  return {
    pressureDevices,
    measureDevices,
    connectPressureDevice,
    disconnectPressureDevice,
    connectMeasureDevice,
    disconnectMeasureDevice
  }
})
