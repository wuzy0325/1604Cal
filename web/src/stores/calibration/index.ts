import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export enum CalibrationStep {
  DEVICE_CONNECT = 0,
  CHANNEL_SELECT = 1,
  START_CALIBRATION = 2,
  DATA_COLLECTION = 3,
  DATA_FITTING = 4,
  COMPLETED = 5
}

export interface PressurePoint {
  id: string
  index: number
  targetPressure: number
  status: 'pending_press' | 'pending_confirm' | 'pending_collect' | 'completed'
  collectedData?: number[]
}

export const useCalibrationStore = defineStore('calibration', () => {
  // State
  const currentStep = ref(CalibrationStep.DEVICE_CONNECT)
  const device1604Connected = ref(false)
  const pressDeviceConnected = ref(false)
  const selectedChannels = ref<number[]>([])
  const pressurePoints = ref<PressurePoint[]>([])
  
  // Getters
  const channelsSelected = computed(() => selectedChannels.value.length > 0)
  const hasCollectedData = computed(() => 
    pressurePoints.value.some(p => p.status === 'completed')
  )
  
  // Actions
  const setStep = (step: CalibrationStep) => {
    currentStep.value = step
  }
  
  const toggleDevice1604 = () => {
    device1604Connected.value = !device1604Connected.value
  }
  
  const togglePressDevice = () => {
    pressDeviceConnected.value = !pressDeviceConnected.value
  }
  
  const setSelectedChannels = (channels: number[]) => {
    selectedChannels.value = channels
  }
  
  const addPressurePoint = (point: PressurePoint) => {
    pressurePoints.value.push(point)
  }
  
  return {
    currentStep,
    device1604Connected,
    pressDeviceConnected,
    selectedChannels,
    pressurePoints,
    channelsSelected,
    hasCollectedData,
    setStep,
    toggleDevice1604,
    togglePressDevice,
    setSelectedChannels,
    addPressurePoint
  }
})