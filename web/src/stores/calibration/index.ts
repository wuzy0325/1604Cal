import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { triggerSessionAction } from '@/services/apiClient'
import { useMeasurementDeviceStore } from '@/stores/measurement/deviceStore'

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

export interface CalibrationParams {
  minValue: number
  maxValue: number
  points: number
  precision: number
  averageCount: number
  stableTime: number
  precisionLevel: string
}

export const useCalibrationStore = defineStore('calibration', () => {
  // 引用设备store
  const deviceStore = useMeasurementDeviceStore()

  // State
  const currentStep = ref(CalibrationStep.DEVICE_CONNECT)
  const selectedChannels = ref<number[]>([])
  const pressurePoints = ref<PressurePoint[]>([])
  const calibrationParams = ref<CalibrationParams>({
    minValue: 0,
    maxValue: 100,
    points: 10,
    precision: 2,
    averageCount: 5,
    stableTime: 3,
    precisionLevel: '0.05'
  })
  const isCollecting = ref(false)
  const currentCollectingPoint = ref(0)

  // Getters - 从设备store获取设备连接状态
  const device1604Connected = computed(() => {
    return deviceStore.measureDevices.some(d => d.status === 'connected')
  })

  const pressDeviceConnected = computed(() => {
    return deviceStore.pressureDevices.some(d => d.status === 'connected')
  })

  const channelsSelected = computed(() => selectedChannels.value.length > 0)
  const hasCollectedData = computed(() => 
    pressurePoints.value.some(p => p.status === 'completed')
  )

  const canStartCalibration = computed(() => 
    device1604Connected.value && 
    pressDeviceConnected.value && 
    channelsSelected.value
  )

  // Actions
  const setStep = (step: CalibrationStep) => {
    currentStep.value = step
  }

  // 连接1604设备
  const connectDevice1604 = async (deviceId: string) => {
    try {
      await deviceStore.connectMeasureDevice(deviceId)
      if (device1604Connected.value && pressDeviceConnected.value) {
        setStep(CalibrationStep.CHANNEL_SELECT)
      }
    } catch (error) {
      console.error('连接1604设备失败:', error)
    }
  }

  // 断开1604设备
  const disconnectDevice1604 = async (deviceId: string) => {
    try {
      await deviceStore.disconnectMeasureDevice(deviceId)
      if (!device1604Connected.value) {
        setStep(CalibrationStep.DEVICE_CONNECT)
      }
    } catch (error) {
      console.error('断开1604设备失败:', error)
    }
  }

  // 连接打压设备
  const connectPressDevice = async (deviceId: string) => {
    try {
      await deviceStore.connectPressureDevice(deviceId)
      if (device1604Connected.value && pressDeviceConnected.value) {
        setStep(CalibrationStep.CHANNEL_SELECT)
      }
    } catch (error) {
      console.error('连接打压设备失败:', error)
    }
  }

  // 断开打压设备
  const disconnectPressDevice = async (deviceId: string) => {
    try {
      await deviceStore.disconnectPressureDevice(deviceId)
      if (!pressDeviceConnected.value) {
        setStep(CalibrationStep.DEVICE_CONNECT)
      }
    } catch (error) {
      console.error('断开打压设备失败:', error)
    }
  }

  const setSelectedChannels = (channels: number[]) => {
    selectedChannels.value = channels
    if (channels.length > 0 && currentStep.value === CalibrationStep.CHANNEL_SELECT) {
      // 保持当前步骤或前进
    }
  }

  // 生成压力点
  const generatePressurePoints = () => {
    const { minValue, maxValue, points } = calibrationParams.value
    const step = (maxValue - minValue) / (points - 1)
    
    pressurePoints.value = Array.from({ length: points }, (_, i) => ({
      id: `point-${i}`,
      index: i + 1,
      targetPressure: minValue + step * i,
      status: i === 0 ? 'pending_press' : 'pending_press'
    }))
    
    ElMessage.success(`已生成 ${points} 个压力点`)
  }

  // 添加压力点
  const addPressurePoint = (point: Omit<PressurePoint, 'id'>) => {
    pressurePoints.value.push({
      ...point,
      id: crypto.randomUUID()
    })
  }

  // 删除压力点
  const removePressurePoint = (index: number) => {
    pressurePoints.value.splice(index, 1)
  }

  // 更新压力点状态
  const updatePointStatus = (pointId: string, status: PressurePoint['status']) => {
    const point = pressurePoints.value.find(p => p.id === pointId)
    if (point) {
      point.status = status
    }
  }

  // 开始校准
  const startCalibration = async () => {
    if (!canStartCalibration.value) {
      ElMessage.warning('请先连接设备并选择通道')
      return
    }

    try {
      await triggerSessionAction('start')
      isCollecting.value = true
      setStep(CalibrationStep.DATA_COLLECTION)
      ElMessage.success('校准已开始')
    } catch (error) {
      console.error('开始校准失败:', error)
      ElMessage.error('开始校准失败')
    }
  }

  // 暂停校准
  const pauseCalibration = async () => {
    try {
      await triggerSessionAction('pause')
      isCollecting.value = false
      ElMessage.info('校准已暂停')
    } catch (error) {
      console.error('暂停校准失败:', error)
    }
  }

  // 恢复校准
  const resumeCalibration = async () => {
    try {
      await triggerSessionAction('resume')
      isCollecting.value = true
      ElMessage.success('校准已恢复')
    } catch (error) {
      console.error('恢复校准失败:', error)
    }
  }

  // 停止校准
  const stopCalibration = async () => {
    try {
      await triggerSessionAction('stop')
      isCollecting.value = false
      setStep(CalibrationStep.START_CALIBRATION)
      ElMessage.info('校准已停止')
    } catch (error) {
      console.error('停止校准失败:', error)
    }
  }

  // 数据拟合
  const fitData = () => {
    if (!hasCollectedData.value) {
      ElMessage.warning('没有可拟合的数据')
      return
    }
    
    setStep(CalibrationStep.DATA_FITTING)
    ElMessage.success('数据拟合完成')
    
    // 模拟拟合完成后进入完成状态
    setTimeout(() => {
      setStep(CalibrationStep.COMPLETED)
    }, 1000)
  }

  // 结束校准
  const endCalibration = () => {
    isCollecting.value = false
    currentCollectingPoint.value = 0
    setStep(CalibrationStep.DEVICE_CONNECT)
    selectedChannels.value = []
    pressurePoints.value = []
    ElMessage.success('校准流程已重置')
  }

  // 采集数据
  const collectData = async (pointId: string) => {
    const point = pressurePoints.value.find(p => p.id === pointId)
    if (!point) return

    try {
      point.status = 'pending_collect'
      // 这里应该调用实际的采集API
      // 模拟采集
      await new Promise(resolve => setTimeout(resolve, 1000))
      
      point.collectedData = selectedChannels.value.map(() => 
        point.targetPressure + (Math.random() - 0.5) * 0.01
      )
      point.status = 'completed'
      
      ElMessage.success(`压力点 ${point.index} 采集完成`)
    } catch (error) {
      console.error('采集数据失败:', error)
      ElMessage.error('采集数据失败')
    }
  }

  // 打压
  const pressurize = async (pointId: string) => {
    const point = pressurePoints.value.find(p => p.id === pointId)
    if (!point) return

    try {
      point.status = 'pending_confirm'
      ElMessage.info(`正在打压至 ${point.targetPressure} kPa`)
      // 这里应该调用实际的打压API
    } catch (error) {
      console.error('打压失败:', error)
      ElMessage.error('打压失败')
    }
  }

  // 确认压力
  const confirmPressure = (pointId: string) => {
    updatePointStatus(pointId, 'pending_collect')
    ElMessage.success('压力已确认，可以进行采集')
  }

  return {
    // State
    currentStep,
    selectedChannels,
    pressurePoints,
    calibrationParams,
    isCollecting,
    currentCollectingPoint,
    // Getters
    device1604Connected,
    pressDeviceConnected,
    channelsSelected,
    hasCollectedData,
    canStartCalibration,
    // Actions
    setStep,
    connectDevice1604,
    disconnectDevice1604,
    connectPressDevice,
    disconnectPressDevice,
    setSelectedChannels,
    generatePressurePoints,
    addPressurePoint,
    removePressurePoint,
    updatePointStatus,
    startCalibration,
    pauseCalibration,
    resumeCalibration,
    stopCalibration,
    fitData,
    endCalibration,
    collectData,
    pressurize,
    confirmPressure
  }
})
