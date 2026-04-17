import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import {
  triggerSessionAction,
  fetchSessionState,
  setCalibrationDevices,
  setCalibrationMeasureDevice,
  setCalibrationConfig,
  setCalibrationChannels,
  generatePressurePoints as apiGeneratePoints,
  getPressurePoints as apiGetPoints,
  pressurize as apiPressurize,
  collectData as apiCollectData,
  fitData as apiFitData,
  readCurrentPressure,
  readStability,
  readMeasureData,
  readCalibrationValve,
  setCalibrationValve as apiSetCalibrationValve,
  readCalibrationMeasureUnit,
  setCalibrationMeasureUnit as apiSetCalibrationMeasureUnit,
  readCalibrationDeviceInfo,
  resetCalibrationDevice,
  multipressRegister,
  multipressUnregister,
  type SessionState
} from '@/services/apiClient'
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
  status: 'pending' | 'pressurizing' | 'stabilizing' | 'collecting' | 'completed' | 'error'
  collectedData?: number[]
  actualPressure?: number
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

// 会话状态到校准步骤的映射
function sessionStateToStep(state: SessionState): CalibrationStep {
  switch (state) {
    case 'idle':
    case 'stopped':
      return CalibrationStep.DEVICE_CONNECT
    case 'ready':
      return CalibrationStep.CHANNEL_SELECT
    case 'pressurizing':
    case 'stabilizing':
    case 'collecting':
    case 'point_done':
    case 'await_manual_collect':
    case 'await_alarm_resolution':
      return CalibrationStep.DATA_COLLECTION
    case 'fitting':
      return CalibrationStep.DATA_FITTING
    case 'completed':
      return CalibrationStep.COMPLETED
    case 'paused':
    case 'recovering':
    case 'error':
      return CalibrationStep.DATA_COLLECTION
    default:
      return CalibrationStep.DEVICE_CONNECT
  }
}

// 判断会话状态是否为"运行中"
function isSessionRunning(state: SessionState): boolean {
  return ['pressurizing', 'stabilizing', 'collecting', 'point_done', 'fitting', 'await_manual_collect', 'await_alarm_resolution', 'recovering'].includes(state)
}

export const useCalibrationStore = defineStore('calibration', () => {
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
  const sessionState = ref<SessionState>('idle')
  const currentPressure = ref(0)
  const isStable = ref(false)
  const channelData = ref<number[]>([])
  const valveStatus = ref<string>('')
  const measureUnit = ref<string>('')
  const deviceInfo = ref<Record<string, string>>({})

  // Getters
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

  const isRunning = computed(() => isSessionRunning(sessionState.value))

  // Actions
  const setStep = (step: CalibrationStep) => {
    currentStep.value = step
  }

  // 同步会话状态
  const syncSessionState = (state: SessionState) => {
    sessionState.value = state
    currentStep.value = sessionStateToStep(state)
    isCollecting.value = isSessionRunning(state)
  }

  // 从后端获取会话状态
  const fetchCurrentSessionState = async () => {
    try {
      const data = await fetchSessionState()
      syncSessionState(data.state)
    } catch (error) {
      console.error('获取会话状态失败:', error)
    }
  }

  const sleep = (ms: number) => new Promise<void>((resolve) => {
    setTimeout(resolve, ms)
  })

  // 连接1604设备
  const connectDevice1604 = async (deviceId: string) => {
    try {
      const connected = await deviceStore.connectMeasureDevice(deviceId)
      if (!connected) {
        return
      }

      // 连接成功后立即绑定驱动到校准服务，使其能读取阀门/单位/设备信息
      try {
        await setCalibrationMeasureDevice(deviceId)
      } catch (err) {
        ElMessage.error('绑定计量设备到校准服务失败，无法读取阀门/单位信息')
        console.error('setCalibrationMeasureDevice failed:', err)
        return
      }

      // 读取设备信息、阀门状态和单位（连接后增加重试，避免设备刚建链时读数失败）
      const loaded = await refreshDeviceInfo({ retries: 3, retryDelayMs: 500 })
      if (!loaded) {
        ElMessage.warning('设备已连接，但阀门/单位信息读取失败，请稍后重试')
      }

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

  // 连接打压设备：通过 multipress 服务注册（创建驱动 + TCP连接 + 注册到压力控制模块）
  const connectPressDevice = async (deviceId: string) => {
    try {
      await multipressRegister(deviceId)
      // multipress 服务不更新 DeviceManager 状态，需手动同步前端 store
      deviceStore.updateDeviceStatus(deviceId, 'connected')
      if (device1604Connected.value && pressDeviceConnected.value) {
        setStep(CalibrationStep.CHANNEL_SELECT)
      }
    } catch (error) {
      console.error('连接打压设备失败:', error)
      ElMessage.error('连接打压设备失败')
    }
  }

  // 断开打压设备：通过 multipress 服务注销（停止控制 + 断开TCP + 移除注册）
  const disconnectPressDevice = async (deviceId: string) => {
    try {
      await multipressUnregister(deviceId)
      deviceStore.updateDeviceStatus(deviceId, 'disconnected')
      if (!pressDeviceConnected.value) {
        setStep(CalibrationStep.DEVICE_CONNECT)
      }
    } catch (error) {
      console.error('断开打压设备失败:', error)
      ElMessage.error('断开打压设备失败')
    }
  }

  // 设置校准设备（通知后端）
  const setDevices = async (measureDeviceId: string, pressureDeviceId: string) => {
    try {
      await setCalibrationDevices({ measureDeviceId, pressureDeviceId })
    } catch (error) {
      console.error('设置校准设备失败:', error)
      ElMessage.error('设置校准设备失败')
    }
  }

  // 设置采集通道
  const setSelectedChannels = async (channels: number[]) => {
    selectedChannels.value = channels
    try {
      await setCalibrationChannels(channels)
    } catch (error) {
      console.error('设置通道失败:', error)
    }
  }

  // 生成压力点
  const generatePressurePoints = async (opts?: { controlMode?: string; pressureMode?: string }) => {
    try {
      // 先设置配置
      await setCalibrationConfig({
        channels: selectedChannels.value,
        pressurePoints: calibrationParams.value.points,
        averageCount: calibrationParams.value.averageCount,
        minPressure: calibrationParams.value.minValue,
        maxPressure: calibrationParams.value.maxValue,
        stableWaitMs: calibrationParams.value.stableTime * 1000,
        controlMode: (opts?.controlMode as 'auto' | 'manual') || undefined,
        pressureMode: (opts?.pressureMode as 'single' | 'return') || undefined
      })

      const points = await apiGeneratePoints()
      pressurePoints.value = points.map(p => ({
        id: `point-${p.index}`,
        index: p.index,
        targetPressure: p.targetPressure,
        status: p.status as PressurePoint['status'],
        collectedData: p.collectedData,
        actualPressure: p.actualPressure
      }))

      ElMessage.success(`已生成 ${points.length} 个压力点`)
    } catch (error) {
      console.error('生成压力点失败:', error)
      ElMessage.error('生成压力点失败')
    }
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
      // 设置后端设备
      const measureDev = deviceStore.measureDevices.find(d => d.status === 'connected')
      const pressureDev = deviceStore.pressureDevices.find(d => d.status === 'connected')
      if (measureDev && pressureDev) {
        await setDevices(measureDev.id, pressureDev.id)
      }

      const data = await triggerSessionAction('start')
      syncSessionState(data.state)
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
      const data = await triggerSessionAction('pause')
      syncSessionState(data.state)
      isCollecting.value = false
      ElMessage.info('校准已暂停')
    } catch (error) {
      console.error('暂停校准失败:', error)
    }
  }

  // 恢复校准
  const resumeCalibration = async () => {
    try {
      const data = await triggerSessionAction('resume')
      syncSessionState(data.state)
      isCollecting.value = true
      ElMessage.success('校准已恢复')
    } catch (error) {
      console.error('恢复校准失败:', error)
    }
  }

  // 停止校准
  const stopCalibration = async () => {
    try {
      const data = await triggerSessionAction('stop')
      syncSessionState(data.state)
      isCollecting.value = false
      setStep(CalibrationStep.START_CALIBRATION)
      ElMessage.info('校准已停止')
    } catch (error) {
      console.error('停止校准失败:', error)
    }
  }

  // 打压
  const pressurize = async (pointId: string) => {
    const point = pressurePoints.value.find(p => p.id === pointId)
    if (!point) return

    try {
      point.status = 'pressurizing'
      await apiPressurize(point.index)

      // 打压完成后刷新压力点状态
      const points = await apiGetPoints()
      const updatedPoint = points.find(p => p.index === point.index)
      if (updatedPoint) {
        point.status = updatedPoint.status as PressurePoint['status']
        point.actualPressure = updatedPoint.actualPressure
      } else {
        point.status = 'stabilizing'
      }

      ElMessage.success(`压力点 ${point.index} 打压完成，压力已稳定`)
    } catch (error) {
      console.error('打压失败:', error)
      point.status = 'error'
      ElMessage.error('打压失败')
    }
  }

  // 确认压力
  const confirmPressure = (pointId: string) => {
    updatePointStatus(pointId, 'collecting')
    ElMessage.success('压力已确认，可以进行采集')
  }

  // 采集数据
  const collectData = async (pointId: string) => {
    const point = pressurePoints.value.find(p => p.id === pointId)
    if (!point) return

    try {
      point.status = 'collecting'
      const data = await apiCollectData(point.index)

      point.collectedData = data
      point.status = 'completed'

      ElMessage.success(`压力点 ${point.index} 采集完成`)
    } catch (error) {
      console.error('采集数据失败:', error)
      point.status = 'error'
      ElMessage.error('采集数据失败')
    }
  }

  // 数据拟合
  const fitData = async () => {
    if (!hasCollectedData.value) {
      ElMessage.warning('没有可拟合的数据')
      return
    }

    try {
      setStep(CalibrationStep.DATA_FITTING)
      await apiFitData()
      setStep(CalibrationStep.COMPLETED)
      sessionState.value = 'completed'
      ElMessage.success('数据拟合完成')
    } catch (error) {
      console.error('拟合失败:', error)
      ElMessage.error('数据拟合失败')
    }
  }

  // 结束校准（先通知后端停止会话，再重置前端状态）
  const endCalibration = async () => {
    // 如果会话还在运行，先通知后端停止
    if (isSessionRunning(sessionState.value)) {
      try {
        await triggerSessionAction('stop')
      } catch (error) {
        console.error('停止后端会话失败:', error)
      }
    }
    isCollecting.value = false
    currentCollectingPoint.value = 0
    setStep(CalibrationStep.DEVICE_CONNECT)
    sessionState.value = 'idle'
    selectedChannels.value = []
    pressurePoints.value = []
    ElMessage.success('校准流程已重置')
  }

  // 重置采集数据（计量模块专用，仅重置测点状态，保留配置）
  const resetCollection = () => {
    pressurePoints.value = pressurePoints.value.map(p => ({
      ...p,
      status: 'pending' as PressurePoint['status'],
      collectedData: undefined,
      actualPressure: undefined
    }))
    isCollecting.value = false
    currentCollectingPoint.value = 0
    sessionState.value = 'idle'
    setStep(CalibrationStep.CHANNEL_SELECT)
    ElMessage.info('采集数据已重置')
  }

  // 读取实时压力
  const refreshPressure = async () => {
    try {
      currentPressure.value = await readCurrentPressure()
    } catch {
      // 静默失败
    }
  }

  // 读取稳定状态
  const refreshStability = async () => {
    try {
      isStable.value = await readStability()
    } catch {
      // 静默失败
    }
  }

  // 读取计量设备数据
  const refreshMeasureData = async () => {
    try {
      channelData.value = await readMeasureData()
    } catch {
      // 静默失败
    }
  }

  // 刷新设备信息（阀门状态、单位、设备信息）
  const refreshDeviceInfo = async (
    options: { retries?: number; retryDelayMs?: number } = {}
  ): Promise<boolean> => {
    const retries = Math.max(1, options.retries ?? 1)
    const retryDelayMs = Math.max(0, options.retryDelayMs ?? 0)
    let valveReady = false
    let unitReady = false

    const tryRead = async <T>(reader: () => Promise<T>): Promise<PromiseSettledResult<T>> => {
      try {
        return { status: 'fulfilled', value: await reader() }
      } catch (error) {
        return { status: 'rejected', reason: error }
      }
    }

    for (let attempt = 1; attempt <= retries; attempt++) {
      const [valve, unit] = await Promise.allSettled([
        readCalibrationValve(),
        readCalibrationMeasureUnit()
      ])

      if (valve.status === 'fulfilled') {
        valveStatus.value = valve.value
        valveReady = true
      }
      if (unit.status === 'fulfilled') {
        measureUnit.value = unit.value
        unitReady = true
      }
      // 连接阶段只要求阀门和单位可读；且允许在不同重试轮次分别成功。
      if (valveReady && unitReady) {
        // 设备信息是增强信息，避免其失败拖累阀门/单位可用性判定。
        const info = await tryRead(readCalibrationDeviceInfo)
        if (info.status === 'fulfilled') {
          deviceInfo.value = info.value
        }
        return true
      }

      if (attempt < retries && retryDelayMs > 0) {
        await sleep(retryDelayMs)
      }
    }

    return false
  }

  // 读取阀门状态
  const refreshValveStatus = async () => {
    try {
      valveStatus.value = await readCalibrationValve()
    } catch {
      // 静默失败
    }
  }

  // 设置阀门状态
  const setValveStatus = async (status: string) => {
    try {
      await apiSetCalibrationValve(status)
      valveStatus.value = status
      ElMessage.success(status === 'calibration' ? '阀门已切换到校准模式' : '阀门已切换到测量模式')
    } catch (error) {
      console.error('设置阀门状态失败:', error)
      ElMessage.error('设置阀门状态失败')
    }
  }

  // 读取计量设备单位
  const refreshMeasureUnit = async () => {
    try {
      measureUnit.value = await readCalibrationMeasureUnit()
    } catch {
      // 静默失败
    }
  }

  // 设置计量设备单位
  const setMeasureUnit = async (unit: string) => {
    try {
      await apiSetCalibrationMeasureUnit(unit)
      measureUnit.value = await readCalibrationMeasureUnit()
      ElMessage.success(`单位已切换为 ${measureUnit.value || unit}`)
    } catch (error) {
      console.error('设置计量设备单位失败:', error)
      ElMessage.error('设置计量设备单位失败')
    }
  }

  // 复位计量设备
  const resetDevice = async () => {
    try {
      await resetCalibrationDevice()
      ElMessage.success('设备已复位')
    } catch (error) {
      console.error('复位设备失败:', error)
      ElMessage.error('复位设备失败')
    }
  }

  return {
    // State
    currentStep,
    selectedChannels,
    pressurePoints,
    calibrationParams,
    isCollecting,
    currentCollectingPoint,
    sessionState,
    currentPressure,
    isStable,
    channelData,
    valveStatus,
    measureUnit,
    deviceInfo,
    // Getters
    device1604Connected,
    pressDeviceConnected,
    channelsSelected,
    hasCollectedData,
    canStartCalibration,
    isRunning,
    // Actions
    setStep,
    syncSessionState,
    fetchCurrentSessionState,
    connectDevice1604,
    disconnectDevice1604,
    connectPressDevice,
    disconnectPressDevice,
    setDevices,
    setSelectedChannels,
    generatePressurePoints,
    addPressurePoint,
    removePressurePoint,
    updatePointStatus,
    startCalibration,
    pauseCalibration,
    resumeCalibration,
    stopCalibration,
    pressurize,
    confirmPressure,
    collectData,
    fitData,
    endCalibration,
    resetCollection,
    refreshPressure,
    refreshStability,
    refreshMeasureData,
    refreshDeviceInfo,
    refreshValveStatus,
    setValveStatus,
    refreshMeasureUnit,
    setMeasureUnit,
    resetDevice
  }
})
