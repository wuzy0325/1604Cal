import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { ActionResult } from '@/types/api'
import {
  triggerSessionAction,
  fetchSessionState,
  setCalibrationConfig,
  fitData as apiFitData,
  resolveAlarm as apiResolveAlarm,
  type AlarmConfigPayload
} from "@/api/calibration"
import { type SessionState, ControlMode } from "@/types/calibration"
import { useDeviceInventoryStore } from '@/stores/device/inventoryStore'
import { usePressurePointStore } from './pressurePoints'
import { useDeviceControlStore } from './deviceControl'
import { sessionStateToStep, isSessionRunning } from '@/composables/useCalibrationFlow'
import { useCalibrationConfig } from '@/composables/useCalibrationConfig'
import { CalibrationStep } from './types'
import { fetchUnitConsistency } from '@/api/device'

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
  const controlMode = ref<ControlMode>(ControlMode.Auto)
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

  const connectDevice1604 = async (deviceId: string): Promise<ActionResult> => {
    const result = await deviceControlStore.connectDevice1604(deviceId)
    if (result.ok && deviceControlStore.device1604Connected) {
      setStep(CalibrationStep.CHANNEL_SELECT)
    }
    return result
  }

  const disconnectDevice1604 = async (deviceId: string): Promise<ActionResult> => {
    const result = await deviceControlStore.disconnectDevice1604(deviceId)
    if (!deviceControlStore.device1604Connected) {
      setStep(CalibrationStep.DEVICE_CONNECT)
    }
    return result
  }

  const connectPressDevice = async (deviceId: string): Promise<ActionResult> => {
    const result = await deviceControlStore.connectPressDevice(deviceId)
    if (deviceControlStore.device1604Connected && deviceControlStore.pressDeviceConnected) {
      setStep(CalibrationStep.CHANNEL_SELECT)
    }
    return result
  }

  const disconnectPressDevice = async (deviceId: string): Promise<ActionResult> => {
    const result = await deviceControlStore.disconnectPressDevice(deviceId)
    if (!deviceControlStore.pressDeviceConnected) {
      setStep(CalibrationStep.DEVICE_CONNECT)
    }
    return result
  }

  const setSelectedChannels = calibrationConfig.setSelectedChannels

  const generatePressurePoints = async (opts?: { controlMode?: string; pressureMode?: string; silent?: boolean }): Promise<ActionResult> => {
    const activeControlMode: ControlMode = opts?.controlMode === ControlMode.Manual ? ControlMode.Manual : controlMode.value
    return pressurePointStore.generatePressurePoints({
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

  const startCalibration = async (opts?: { controlMode?: string }): Promise<ActionResult> => {
    const activeControlMode: ControlMode = opts?.controlMode === ControlMode.Manual ? ControlMode.Manual : controlMode.value
    controlMode.value = activeControlMode

    if (!canStartCalibration.value) {
      const missing: string[] = []
      if (!device1604Connected.value) missing.push('连接计量设备')
      if (!channelsSelected.value) missing.push('选择通道')
      if (enforceValveCalibrationGate && !valveReady.value) missing.push('将阀门切换到校准状态')
      return { ok: false, error: 'MISSING_REQUIREMENTS', detail: `请先${missing.join('并')}` }
    }
    // 自动模式额外校验打压设备
    if (activeControlMode === ControlMode.Auto && !pressDeviceConnected.value) {
      return { ok: false, error: 'MISSING_PRESS_DEVICE', detail: '自动模式需要连接打压设备' }
    }

    // 自动模式校验采集设备和打压设备单位一致
    if (activeControlMode === ControlMode.Auto) {
      try {
        const unitCheck = await fetchUnitConsistency()
        if (!unitCheck.consistent) {
          return { ok: false, error: 'UNIT_MISMATCH', detail: '采集设备与打压设备压力单位不一致，请统一单位后再开始标定' }
        }
      } catch {
        return { ok: false, error: 'UNIT_CHECK_FAILED', detail: '无法检查设备单位一致性，请确认设备连接正常' }
      }
    }

    // 保存用户手动编辑的目标压力值，避免重新生成时被覆盖
    const savedTargets = new Map(
      pressurePointStore.pressurePoints.map(p => [p.index, p.targetPressure])
    )

    // 重新生成压力点以初始化后端内存中的点位
    const pointsResult = await pressurePointStore.generatePressurePoints({
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
    if (!pointsResult.ok || pressurePointStore.pressurePoints.length === 0) {
      return { ok: false, error: 'POINTS_NOT_READY', detail: '开始标定失败：压力点未就绪' }
    }

    // 恢复用户手动编辑的目标压力值
    const restoreResults = await Promise.all(
      pressurePointStore.pressurePoints.map(async (point) => {
        const edited = savedTargets.get(point.index)
        if (edited === undefined || Math.abs(edited - point.targetPressure) <= 0.001) return null
        return pressurePointStore.updateTargetPressure(point.id, edited)
      })
    )
    const failures = restoreResults.filter(r => r !== null && !r.ok)
    if (failures.length > 0) {
      console.warn(`${failures.length} 个压力点的编辑值恢复失败`, failures)
    }

    return pushCalibrationConfigAndStart(activeControlMode)
  }

  const pushCalibrationConfigAndStart = async (controlMode: ControlMode): Promise<ActionResult> => {
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
        controlMode,
        precision: calibrationParams.value.precision,
        precisionLevel: Number(calibrationParams.value.precisionLevel) || 0.05,
        pressureMode: calibrationParams.value.pressureMode
      })
      const data = await triggerSessionAction('start')
      syncSessionState(data.state)
      isCollecting.value = true
      setStep(CalibrationStep.DATA_COLLECTION)
      return { ok: true }
    } catch (error) {
      console.error('开始标定失败:', error)
      const detail = error instanceof Error ? error.message : String(error)
      return { ok: false, error: 'START_FAILED', detail: `开始标定失败: ${detail}` }
    }
  }

  const withSessionAction = async (action: 'start' | 'pause' | 'resume' | 'stop', onSuccess?: () => void): Promise<ActionResult> => {
    try {
      const data = await triggerSessionAction(action)
      syncSessionState(data.state)
      onSuccess?.()
      return { ok: true }
    } catch (error) {
      console.error(`${action} 失败:`, error)
      return { ok: false, error: `${action.toUpperCase()}_FAILED`, detail: String(error) }
    }
  }

  const pauseCalibration = () => withSessionAction('pause', () => { isCollecting.value = false })

  const resumeCalibration = () => withSessionAction('resume', () => { isCollecting.value = true })

  const stopCalibration = () => withSessionAction('stop', () => { isCollecting.value = false; setStep(CalibrationStep.START_CALIBRATION) })

  const resolveAlarm = async (decision: 'continue' | 'skip' | 'recollect' | 'stop'): Promise<ActionResult> => {
    try {
      await apiResolveAlarm(decision)
      return { ok: true }
    } catch (error) {
      console.error('报警处理失败:', error)
      return { ok: false, error: 'ALARM_RESOLVE_FAILED', detail: '报警处理失败' }
    }
  }

  const canOperateCurrentPoint = () => {
    return sessionState.value !== 'idle' && sessionState.value !== 'stopped' && sessionState.value !== 'completed'
  }

  const pressurize = async (pointId: string): Promise<ActionResult> => {
    if (!canOperateCurrentPoint()) {
      return { ok: false, error: 'NOT_RUNNING', detail: '请先开始标定流程' }
    }

    if (controlMode.value === ControlMode.Manual && !pressDeviceConnected.value) {
      return { ok: false, error: 'NOT_CONNECTED', detail: '手动模式且未连接打压设备，请先确认压力到位' }
    }
    return pressurePointStore.pressurize(pointId)
  }

  const fitData = async (): Promise<ActionResult> => {
    if (!hasCollectedData.value) {
      return { ok: false, error: 'NO_DATA', detail: '没有可拟合的数据' }
    }
    try {
      setStep(CalibrationStep.DATA_FITTING)
      const result = await apiFitData()
      pressurePointStore.fittingResult = result
      setStep(CalibrationStep.COMPLETED)
      sessionState.value = 'completed'
      return { ok: true }
    } catch (error) {
      console.error('拟合失败:', error)
      return { ok: false, error: 'FIT_FAILED', detail: '数据拟合失败' }
    }
  }

  const endCalibration = async (): Promise<ActionResult> => {
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
    return { ok: true }
  }

  const resetCollection = (): ActionResult => {
    pressurePointStore.resetCollection()
    isCollecting.value = false
    currentCollectingPoint.value = 0
    sessionState.value = 'idle'
    setStep(CalibrationStep.CHANNEL_SELECT)
    return { ok: true }
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
    updateTargetPressure: pressurePointStore.updateTargetPressure,
    removePressurePoint: pressurePointStore.removePressurePoint,
    updatePointStatus: pressurePointStore.updatePointStatus,
    startCalibration,
    pauseCalibration,
    resumeCalibration,
    stopCalibration,
    resolveAlarm,
    pressurize,
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
