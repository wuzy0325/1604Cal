import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import {
  triggerSessionAction,
  fetchSessionState,
  setCalibrationConfig,
  fitData as apiFitData,
  resolveAlarm as apiResolveAlarm,
  type AlarmConfigPayload
} from "@/api/calibration"
import type { SessionState } from "@/types/calibration"
import { useDeviceInventoryStore } from '@/stores/device/inventoryStore'
import { usePressurePointStore } from './pressurePoints'
import { useDeviceControlStore } from './deviceControl'
import { sessionStateToStep, isSessionRunning } from '@/composables/useCalibrationFlow'
import { useCalibrationConfig } from '@/composables/useCalibrationConfig'
import { CalibrationStep } from './types'

export { CalibrationStep } from './types'
export type { PressurePoint, CalibrationParams } from './types'

export const useCalibrationStore = defineStore('calibration', () => {
  const deviceStore = useDeviceInventoryStore()
  const pressurePointStore = usePressurePointStore()
  const deviceControlStore = useDeviceControlStore()
  const calibrationConfig = useCalibrationConfig()

  // State (own)
  const currentStep = ref(CalibrationStep.DEVICE_CONNECT)
  const isCollecting = ref(false)
  const currentCollectingPoint = ref(0)
  const sessionState = ref<SessionState>('idle')
  const controlMode = ref<'auto' | 'manual'>('auto')
  const alarmConfig = ref<AlarmConfigPayload>({
    enabled: true,
    precisionThreshold: 5.0,
    soundEnabled: true,
    confirmOnAlarm: true,
    enabledChannels: []
  })

  // State (delegated from composable - these are Refs)
  const selectedChannels = calibrationConfig.selectedChannels
  const calibrationParams = calibrationConfig.calibrationParams

  // Getters
  const device1604Connected = computed(() => deviceControlStore.device1604Connected)
  const pressDeviceConnected = computed(() => deviceControlStore.pressDeviceConnected)
  const channelsSelected = computed(() => selectedChannels.value.length > 0)
  const hasCollectedData = computed(() => pressurePointStore.hasCollectedData)
  const valveReady = computed(() => deviceControlStore.valveStatus === 'calibration')
  const enforceValveCalibrationGate = false

  // 前端阀门门禁统一开关：false 表示放开（联调），true 表示严格门禁。
  const canStartCalibration = computed(() =>
    device1604Connected.value && channelsSelected.value && (!enforceValveCalibrationGate || valveReady.value)
  )
  const isRunning = computed(() => isSessionRunning(sessionState.value))

  // Actions
  const setStep = (step: CalibrationStep) => { currentStep.value = step }

  const syncSessionState = (state: SessionState) => {
    sessionState.value = state
    currentStep.value = sessionStateToStep(state)
    isCollecting.value = isSessionRunning(state)
  }

  const fetchCurrentSessionState = async () => {
    try {
      const data = await fetchSessionState()
      syncSessionState(data.state)
    } catch (error) { console.error('获取会话状态失败:', error) }
  }

  const connectDevice1604 = async (deviceId: string) => {
    const result = await deviceControlStore.connectDevice1604(deviceId)
    if (result && deviceControlStore.device1604Connected) {
      setStep(CalibrationStep.CHANNEL_SELECT)
    }
  }

  const disconnectDevice1604 = async (deviceId: string) => {
    await deviceControlStore.disconnectDevice1604(deviceId)
    if (!deviceControlStore.device1604Connected) {
      setStep(CalibrationStep.DEVICE_CONNECT)
    }
  }

  const connectPressDevice = async (deviceId: string) => {
    await deviceControlStore.connectPressDevice(deviceId)
    if (deviceControlStore.device1604Connected && deviceControlStore.pressDeviceConnected) {
      setStep(CalibrationStep.CHANNEL_SELECT)
    }
  }

  const disconnectPressDevice = async (deviceId: string) => {
    await deviceControlStore.disconnectPressDevice(deviceId)
    if (!deviceControlStore.pressDeviceConnected) {
      setStep(CalibrationStep.DEVICE_CONNECT)
    }
  }

  const setSelectedChannels = calibrationConfig.setSelectedChannels

  const generatePressurePoints = async (opts?: { controlMode?: string; pressureMode?: string; silent?: boolean }) => {
    const activeControlMode: 'auto' | 'manual' = opts?.controlMode === 'manual' ? 'manual' : controlMode.value
    await pressurePointStore.generatePressurePoints({
      ...opts,
      controlMode: activeControlMode,
      channels: selectedChannels.value,
      silent: opts?.silent,
      params: {
        points: calibrationParams.value.points,
        averageCount: calibrationParams.value.averageCount,
        minValue: calibrationParams.value.minValue,
        maxValue: calibrationParams.value.maxValue,
        stableTime: calibrationParams.value.stableTime,
        precision: calibrationParams.value.precision,
        precisionLevel: calibrationParams.value.precisionLevel
      }
    })
  }

  const startCalibration = async (opts?: { controlMode?: string }) => {
    const activeControlMode: 'auto' | 'manual' = opts?.controlMode === 'manual' ? 'manual' : controlMode.value
    controlMode.value = activeControlMode

    if (!canStartCalibration.value) {
      const missing: string[] = []
      if (!device1604Connected.value) missing.push('连接计量设备')
      if (!channelsSelected.value) missing.push('选择通道')
      if (enforceValveCalibrationGate && !valveReady.value) missing.push('将阀门切换到校准状态')
      ElMessage.warning(`请先${missing.join('并')}`)
      return
    }
    // 自动模式额外校验打压设备
    if (activeControlMode === 'auto' && !pressDeviceConnected.value) {
      ElMessage.warning('自动模式需要连接打压设备')
      return
    }

    // 对齐旧模块流程：每次开始标定前都重新生成后端压力点，
    // 避免前端本地缓存点位与后端内存点位不一致导致采集失败。
    const pointsReady = await pressurePointStore.generatePressurePoints({
      controlMode: activeControlMode,
      channels: selectedChannels.value,
      params: {
        points: calibrationParams.value.points,
        averageCount: calibrationParams.value.averageCount,
        minValue: calibrationParams.value.minValue,
        maxValue: calibrationParams.value.maxValue,
        stableTime: calibrationParams.value.stableTime,
        precision: calibrationParams.value.precision,
        precisionLevel: calibrationParams.value.precisionLevel
      }
    })
    if (!pointsReady || pressurePointStore.pressurePoints.length === 0) {
      ElMessage.error('开始标定失败：压力点未就绪')
      return
    }

    try {
      const measureDev = deviceStore.measureDevices.find(d => d.status === 'connected')
      const pressureDev = deviceStore.pressureDevices.find(d => d.status === 'connected')
      if (measureDev && pressureDev) {
        await deviceControlStore.setDevices(measureDev.id, pressureDev.id)
      } else if (measureDev) {
        await deviceControlStore.setDevices(measureDev.id, '')
      }
      await setCalibrationConfig({
        channels: selectedChannels.value,
        pressurePoints: calibrationParams.value.points,
        averageCount: calibrationParams.value.averageCount,
        minPressure: calibrationParams.value.minValue,
        maxPressure: calibrationParams.value.maxValue,
        stableWaitMs: calibrationParams.value.stableTime * 1000,
        controlMode: activeControlMode,
        precision: calibrationParams.value.precision,
        precisionLevel: Number(calibrationParams.value.precisionLevel) || 0.05,
        pressureMode: calibrationParams.value.pressureMode
      })
      const data = await triggerSessionAction('start')
      syncSessionState(data.state)
      isCollecting.value = true
      setStep(CalibrationStep.DATA_COLLECTION)
      ElMessage.success('标定已开始')
    } catch (error) {
      console.error('开始标定失败:', error)
      const detail = error instanceof Error ? error.message : String(error)
      ElMessage.error(`开始标定失败: ${detail}`)
    }
  }

  const pauseCalibration = async () => {
    try {
      const data = await triggerSessionAction('pause')
      syncSessionState(data.state)
      isCollecting.value = false
      ElMessage.info('校准已暂停')
    } catch (error) { console.error('暂停校准失败:', error) }
  }

  const resumeCalibration = async () => {
    try {
      const data = await triggerSessionAction('resume')
      syncSessionState(data.state)
      isCollecting.value = true
      ElMessage.success('校准已恢复')
    } catch (error) { console.error('恢复校准失败:', error) }
  }

  const stopCalibration = async () => {
    try {
      const data = await triggerSessionAction('stop')
      syncSessionState(data.state)
      isCollecting.value = false
      setStep(CalibrationStep.START_CALIBRATION)
      ElMessage.info('校准已停止')
    } catch (error) { console.error('停止校准失败:', error) }
  }

  const resolveAlarm = async (decision: 'continue' | 'skip' | 'recollect' | 'stop') => {
    try {
      await apiResolveAlarm(decision)

      const decisionTextMap: Record<typeof decision, string> = {
        continue: '报警已确认，继续流程',
        skip: '已跳过当前测点',
        recollect: '将重新采集当前点',
        stop: '已停止自动采集流程'
      }

      ElMessage.success(decisionTextMap[decision])
    } catch (error) {
      console.error('报警处理失败:', error)
      ElMessage.error('报警处理失败')
    }
  }

  const canOperateCurrentPoint = () => {
    return sessionState.value !== 'idle' && sessionState.value !== 'stopped' && sessionState.value !== 'completed'
  }

  const pressurize = async (pointId: string) => {
    if (!canOperateCurrentPoint()) {
      ElMessage.warning('请先开始标定流程')
      return
    }

    if (controlMode.value === 'manual' && !pressDeviceConnected.value) {
      ElMessage.warning('手动模式且未连接打压设备，请先确认压力到位')
      return
    }
    await pressurePointStore.pressurize(pointId)
  }

  const confirmPressure = (pointId: string) => {
    if (!canOperateCurrentPoint()) {
      ElMessage.warning('请先开始标定流程')
      return
    }

    const point = pressurePointStore.pressurePoints.find(p => p.id === pointId)
    if (!point) {
      return
    }

    // 手动模式下，pending 点可直接确认（操作者自行保证压力已到位）
    if (controlMode.value === 'manual' && point.status === 'pending') {
      point.status = 'stabilizing'
      ElMessage.success('已确认压力到位，可以进行采集')
      return
    }

    pressurePointStore.confirmPressure(pointId)
  }

  const fitData = async () => {
    if (!hasCollectedData.value) {
      ElMessage.warning('没有可拟合的数据')
      return
    }
    try {
      setStep(CalibrationStep.DATA_FITTING)
      const result = await apiFitData()
      pressurePointStore.fittingResult = result
      setStep(CalibrationStep.COMPLETED)
      sessionState.value = 'completed'
      ElMessage.success('数据拟合完成')
    } catch (error) {
      console.error('拟合失败:', error)
      ElMessage.error('数据拟合失败')
    }
  }

  const endCalibration = async () => {
    if (isSessionRunning(sessionState.value)) {
      try { await triggerSessionAction('stop') }
      catch (error) { console.error('停止后端会话失败:', error) }
    } else if (sessionState.value !== 'idle') {
      try { await triggerSessionAction('stop') }
      catch (error) { console.error('结束后端校准失败:', error) }
    }
    isCollecting.value = false
    currentCollectingPoint.value = 0
    setStep(CalibrationStep.DEVICE_CONNECT)
    sessionState.value = 'idle'
    pressurePointStore.resetCollection()
    ElMessage.success('校准流程已重置')
  }

  const resetCollection = () => {
    pressurePointStore.resetCollection()
    isCollecting.value = false
    currentCollectingPoint.value = 0
    sessionState.value = 'idle'
    setStep(CalibrationStep.CHANNEL_SELECT)
    ElMessage.info('采集数据已重置')
  }

  return {
    // State
    currentStep,
    selectedChannels,
    pressurePoints: computed(() => pressurePointStore.pressurePoints),
    calibrationParams,
    isCollecting,
    currentCollectingPoint,
    sessionState,
    controlMode,
    alarmConfig,
    currentPressure: computed(() => deviceControlStore.currentPressure),
    isStable: computed(() => deviceControlStore.isStable),
    channelData: computed(() => deviceControlStore.channelData),
    valveStatus: computed(() => deviceControlStore.valveStatus),
    measureUnit: computed(() => deviceControlStore.measureUnit),
    deviceInfo: computed(() => deviceControlStore.deviceInfo),
    fittingResult: computed(() => pressurePointStore.fittingResult),
    // Getters
    device1604Connected,
    pressDeviceConnected,
    channelsSelected,
    hasCollectedData,
    valveReady,
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
    setSelectedChannels,
    generatePressurePoints,
    addPressurePoint: pressurePointStore.addPressurePoint,
    removePressurePoint: pressurePointStore.removePressurePoint,
    updatePointStatus: pressurePointStore.updatePointStatus,
    startCalibration,
    pauseCalibration,
    resumeCalibration,
    stopCalibration,
    resolveAlarm,
    pressurize,
    confirmPressure,
    collectData: pressurePointStore.collectData,
    fitData,
    endCalibration,
    resetCollection,
    refreshPressure: deviceControlStore.refreshPressure,
    refreshStability: deviceControlStore.refreshStability,
    refreshMeasureData: deviceControlStore.refreshMeasureData,
    refreshDeviceInfo: deviceControlStore.refreshDeviceInfo,
    refreshValveStatus: deviceControlStore.refreshValveStatus,
    setValveStatus: deviceControlStore.setValveStatus,
    refreshMeasureUnit: deviceControlStore.refreshMeasureUnit,
    setMeasureUnit: deviceControlStore.setMeasureUnit,
    resetDevice: deviceControlStore.resetDevice,
    setDevices: deviceControlStore.setDevices
  }
})
