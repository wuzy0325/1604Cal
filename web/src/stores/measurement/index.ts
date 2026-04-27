import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import {
  bindDevices as apiBindDevices,
  bindMeasureDevice as apiBindMeasureDevice,
  readPressure as apiReadPressure,
  readStability as apiReadStability,
  readMeasureData as apiReadMeasureData,
  readValveStatus as apiReadValveStatus,
  setValveStatus as apiSetValveStatus,
  readMeasureUnit as apiReadMeasureUnit,
  setMeasureUnit as apiSetMeasureUnit,
  readDeviceInfo as apiReadDeviceInfo,
  resetDevice as apiResetDevice
} from '@/api/session'
import {
  fetchMeasurementState,
  startMeasurement,
  pauseMeasurement,
  stopMeasurement,
  fetchMeasurementData,
  getMeasurementExportUrl,
  generateMeasurementPoints,
  fetchMeasurementPoints,
  getMeasurementAlarmConfig,
  saveMeasurementAlarmConfig,
  checkMeasurementAlarmPending,
  resolveMeasurementAlarm,
  getMeasurementParamsConfig,
  saveMeasurementParamsConfig,
  autoCollectMeasurement,
  manualPressurizeMeasurement,
  manualCollectMeasurement,
  type MeasurementPoint,
  type MeasurementAlarmConfig,
  type MeasurementParamsPayload
} from '@/api/measurement'
import { createEventStream } from '@/api/client'
import type { StreamEventPayload } from '@/types/api'
import type { MeasurementState, CollectedRow, StabilityUpdate, AlarmData } from './types'

export type { MeasurementState, CollectedRow, StabilityUpdate }

export const useMeasurementStore = defineStore('measurement', () => {
  // ── 状态 ──
  const state = ref<MeasurementState>('idle')
  const rows = ref<CollectedRow[]>([])
  const channels = ref<number[]>([])
  const measureDeviceId = ref('')
  const pressureDeviceId = ref('')

  // 设备实时数据
  const currentPressure = ref(0)
  const isStable = ref(false)
  const channelData = ref<number[]>([])
  const valveStatus = ref('')
  const measureUnit = ref('')
  const deviceInfo = ref<Record<string, string>>({})

  // 稳定性监控状态
  const stabilityState = ref<StabilityUpdate>({
    pointIndex: 0,
    isStable: false,
    isInRange: false,
    currentValue: 0,
    stableDurationMs: 0,
    requiredDurationMs: 0,
    progress: 0
  })

  // 计量参数（UI 直接绑定，与 backend config 映射）
  const measurementParams = ref({
    minPressure: 0,
    maxPressure: 10,
    pointCount: 5,
    precision: 3,
    averageCount: 3,
    stableWaitS: 3,
    precisionLevel: 0.05,
    pressureMode: 'single' as 'single' | 'roundTrip',
    controlMode: 'auto' as 'auto' | 'manual'
  })

  // 计量工作流相关
  const config = ref<MeasurementParamsPayload | null>(null)
  const points = ref<MeasurementPoint[]>([])
  const alarmConfig = ref<MeasurementAlarmConfig>({
    enabled: true,
    enabledChannels: [],
    confirmOnAlarm: false,
    soundEnabled: true,
    threshold: 5.0
  })
  const alarmPending = ref(false)
  const alarmData = ref<AlarmData | null>(null)

  // SSE
  let eventSource: EventSource | null = null

  // ── 计算属性 ──
  const runningStates: MeasurementState[] = ['pressuring', 'stabilizing', 'collecting']
  const startableStates: MeasurementState[] = ['idle', 'completed', 'error']

  const isCollecting = computed(() => state.value === 'collecting')
  const isRunning = computed(() => runningStates.includes(state.value))
  const isStartable = computed(() => startableStates.includes(state.value))
  const isPaused = computed(() => state.value === 'paused')
  const isIdle = computed(() => state.value === 'idle')
  const totalRows = computed(() => rows.value.length)
  const exportUrl = computed(() => getMeasurementExportUrl())
  const deviceBound = computed(() => measureDeviceId.value !== '')

  // ── 设备绑定 ──

  const bindDevices = async (measureDevId: string, pressureDevId: string) => {
    await apiBindDevices(measureDevId, pressureDevId)
    measureDeviceId.value = measureDevId
    pressureDeviceId.value = pressureDevId
  }

  const bindMeasureDevice = async (measureDevId: string) => {
    await apiBindMeasureDevice(measureDevId)
    measureDeviceId.value = measureDevId
  }

  const unbindMeasureDevice = () => {
    measureDeviceId.value = ''
    pressureDeviceId.value = ''
    channelData.value = []
    valveStatus.value = ''
    measureUnit.value = ''
    deviceInfo.value = {}
  }

  const unbindPressureDevice = () => {
    pressureDeviceId.value = ''
  }

  // ── 实时数据读取 ──

  const refreshPressure = async () => {
    try { currentPressure.value = await apiReadPressure() }
    catch { /* 设备未绑定时静默 */ }
  }

  const refreshStability = async () => {
    try { isStable.value = await apiReadStability() }
    catch { /* 静默 */ }
  }

  const refreshMeasureData = async () => {
    try { channelData.value = await apiReadMeasureData() }
    catch { /* 静默 */ }
  }

  const refreshValveStatus = async () => {
    try { valveStatus.value = await apiReadValveStatus() }
    catch { /* 静默 */ }
  }

  const setValveStatus = async (status: string) => {
    await apiSetValveStatus(status)
    valveStatus.value = status
  }

  const refreshMeasureUnit = async () => {
    try { measureUnit.value = await apiReadMeasureUnit() }
    catch { /* 静默 */ }
  }

  const setMeasureUnit = async (unit: string) => {
    await apiSetMeasureUnit(unit)
    measureUnit.value = unit
  }

  const refreshDeviceInfo = async () => {
    try { deviceInfo.value = await apiReadDeviceInfo() }
    catch { /* 静默 */ }
  }

  const resetDevice = async () => {
    await apiResetDevice()
  }

  // ── 采集工作流 ──

  const start = async (selectedChannels: number[]) => {
    if (!deviceBound.value) {
      ElMessage.warning('请先绑定计量设备')
      return
    }
    try {
      channels.value = selectedChannels
      const newState = await startMeasurement(selectedChannels)
      state.value = newState as MeasurementState
      rows.value = []
      ElMessage.success('计量采集已开始')
    } catch (error) {
      const detail = error instanceof Error ? error.message : String(error)
      ElMessage.error(`启动采集失败: ${detail}`)
    }
  }

  const pause = async () => {
    try {
      const newState = await pauseMeasurement()
      state.value = newState as MeasurementState
      ElMessage.info('采集已暂停')
    } catch (error) {
      ElMessage.error('暂停采集失败')
    }
  }

  const stop = async () => {
    try {
      const newState = await stopMeasurement()
      state.value = newState as MeasurementState
      ElMessage.info('采集已停止')
    } catch (error) {
      ElMessage.error('停止采集失败')
    }
  }

  const refreshData = async () => {
    try {
      const resp = await fetchMeasurementData()
      rows.value = resp.rows
    } catch { /* 静默 */ }
  }

  // ── SSE 事件监听 ──

  const setupSSE = () => {
    if (eventSource) return
    eventSource = createEventStream((payload: StreamEventPayload) => {
      switch (payload.type) {
        case 'measurement.state_changed':
          state.value = (payload.data as { state: MeasurementState }).state
          break
        case 'measurement.data_updated': {
          const data = payload.data as { timestamp: string; channels: Record<string, number> }
          rows.value.push({ timestamp: data.timestamp, channels: data.channels })
          break
        }
        case 'measurement.stability.update':
          stabilityState.value = payload.data as StabilityUpdate
          isStable.value = stabilityState.value.isStable
          currentPressure.value = stabilityState.value.currentValue
          break
        case 'measurement.alarm.triggered':
          alarmPending.value = true
          alarmData.value = payload.data as AlarmData
          break
        case 'measurement.alarm.resolved':
          alarmPending.value = false
          alarmData.value = null
          break
        case 'measurement.point.status': {
          const updatedPoint = payload.data as MeasurementPoint
          const idx = points.value.findIndex(p => p.id === updatedPoint.id)
          if (idx >= 0) points.value[idx] = updatedPoint
          break
        }
        case 'measurement.data.collected': {
          const collected = payload.data as { pointIndex: number; channels: number[]; data: number[] }
          const pointIdx = points.value.findIndex(p => p.index === collected.pointIndex)
          if (pointIdx >= 0) {
            points.value[pointIdx] = { ...points.value[pointIdx], collectedData: collected.data, status: 'completed' }
          }
          break
        }
      }
    })
  }

  const teardownSSE = () => {
    if (eventSource) {
      eventSource.close()
      eventSource = null
    }
  }

  const fetchCurrentState = async () => {
    try {
      const s = await fetchMeasurementState()
      state.value = s as MeasurementState
    } catch { /* 静默 */ }
  }

  const syncState = (newState: MeasurementState) => {
    state.value = newState
  }

  // ── 计量工作流 ──

  const loadConfig = async () => {
    try {
      const cfg = await getMeasurementParamsConfig()
      config.value = cfg
      // 同步到 UI 绑定源
      if (cfg) {
        measurementParams.value = {
          minPressure: cfg.minPressure,
          maxPressure: cfg.maxPressure,
          pointCount: cfg.pointCount,
          precision: cfg.precision,
          averageCount: cfg.averageCount,
          stableWaitS: Math.round(cfg.stableDurationMs / 1000),
          precisionLevel: cfg.precisionLevel,
          pressureMode: cfg.pressureMode as 'single' | 'roundTrip',
          controlMode: cfg.controlMode as 'auto' | 'manual'
        }
      }
    } catch { /* 静默 */ }
  }

  const saveConfig = async (params: MeasurementParamsPayload) => {
    await saveMeasurementParamsConfig(params)
    config.value = params
  }

  const loadPoints = async () => {
    try {
      points.value = await fetchMeasurementPoints()
    } catch { /* 静默 */ }
  }

  const generatePoints = async () => {
    try {
      const p = measurementParams.value
      const payload: MeasurementParamsPayload = {
        minPressure: p.minPressure,
        maxPressure: p.maxPressure,
        pointCount: p.pointCount,
        precision: p.precision,
        averageCount: p.averageCount,
        stableDurationMs: p.stableWaitS * 1000,
        precisionLevel: p.precisionLevel,
        pressureMode: p.pressureMode,
        controlMode: p.controlMode
      }
      await saveMeasurementParamsConfig(payload)
      config.value = payload
      points.value = await generateMeasurementPoints()
      ElMessage.success('压力点已生成')
    } catch (error) {
      const detail = error instanceof Error ? error.message : String(error)
      ElMessage.error(`生成压力点失败: ${detail}`)
    }
  }

  const loadAlarmConfig = async () => {
    try {
      alarmConfig.value = await getMeasurementAlarmConfig()
    } catch { /* 静默 */ }
  }

  const saveAlarmConfig = async (cfg: MeasurementAlarmConfig) => {
    await saveMeasurementAlarmConfig(cfg)
    alarmConfig.value = cfg
  }

  const refreshAlarmPending = async () => {
    try {
      alarmPending.value = await checkMeasurementAlarmPending()
    } catch { /* 静默 */ }
  }

  const resolveAlarm = async (decision: 'continue' | 'retry') => {
    await resolveMeasurementAlarm(decision)
    alarmPending.value = false
  }

  // ── 按点采集 ──

  const autoCollect = async () => {
    try {
      const newState = await autoCollectMeasurement()
      state.value = newState as MeasurementState
    } catch (error) {
      const detail = error instanceof Error ? error.message : String(error)
      ElMessage.error(`自动采集失败: ${detail}`)
    }
  }

  const manualPressurize = async (pointIndex: number) => {
    try {
      const newState = await manualPressurizeMeasurement(pointIndex)
      state.value = newState as MeasurementState
    } catch (error) {
      const detail = error instanceof Error ? error.message : String(error)
      ElMessage.error(`手动打压失败: ${detail}`)
    }
  }

  const manualCollect = async (pointIndex: number) => {
    try {
      const newState = await manualCollectMeasurement(pointIndex)
      state.value = newState as MeasurementState
    } catch (error) {
      const detail = error instanceof Error ? error.message : String(error)
      ElMessage.error(`手动采集失败: ${detail}`)
    }
  }

  return {
    // 状态
    state,
    rows,
    channels,
    measureDeviceId,
    pressureDeviceId,
    measurementParams,
    currentPressure,
    isStable,
    channelData,
    valveStatus,
    measureUnit,
    deviceInfo,
    stabilityState,
    // 计算属性
    isCollecting,
    isRunning,
    isStartable,
    isPaused,
    isIdle,
    totalRows,
    exportUrl,
    deviceBound,
    // 设备绑定
    bindDevices,
    bindMeasureDevice,
    unbindMeasureDevice,
    unbindPressureDevice,
    // 实时数据
    refreshPressure,
    refreshStability,
    refreshMeasureData,
    refreshValveStatus,
    setValveStatus,
    refreshMeasureUnit,
    setMeasureUnit,
    refreshDeviceInfo,
    resetDevice,
    // 采集工作流
    start,
    pause,
    stop,
    refreshData,
    // 计量工作流
    config,
    points,
    alarmConfig,
    alarmPending,
    alarmData,
    loadConfig,
    saveConfig,
    loadPoints,
    generatePoints,
    loadAlarmConfig,
    saveAlarmConfig,
    refreshAlarmPending,
    resolveAlarm,
    // 按点采集
    autoCollect,
    manualPressurize,
    manualCollect,
    // SSE
    setupSSE,
    teardownSSE,
    fetchCurrentState,
    syncState
  }
})
