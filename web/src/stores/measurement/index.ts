import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { ActionResult } from '@/types/api'
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
  manualStartMeasurement,
  type MeasurementPoint,
  type MeasurementAlarmConfig,
  type MeasurementParamsPayload,
  type AlarmDecision
} from '@/api/measurement'
import type { MeasurementState, CollectedRow, StabilityUpdate, AlarmData, PrimaryAction, SecondaryAction } from './types'
import { ControlMode, PressureMode } from '@/types/calibration'
import { useMeasurementDeviceStore } from '@/stores/measurement/deviceStore'
import { useGatesStore } from '@/stores/app/gates'
import { fetchUnitConsistency } from '@/api/device'

export type { MeasurementState, CollectedRow, StabilityUpdate }

/**
 * 实时采样行最大保留条数，与后端 measurement.maxCollectedRows 对齐。
 * 超出时丢弃最旧行，防止长时间采集导致响应式 store 与 DOM 无限增长。
 * SSE 追加路径（useMeasurementSync）与全量拉取路径（refreshData）共用此上限。
 */
export const MEASUREMENT_MAX_ROWS = 2000

export const useMeasurementStore = defineStore('measurement', () => {
  // ── 状态 ──
  const state = ref<MeasurementState>('idle')
  const rows = ref<CollectedRow[]>([])
  const channels = ref<number[]>(Array.from({ length: 16 }, (_, i) => i + 1))
  const measureDeviceId = ref('')
  const pressureDeviceId = ref('')

  // 设备实时数据
  const currentPressure = ref(0)
  const isStable = ref(false)
  const channelData = ref<number[]>([])
  const valveStatus = ref('')
  const measureUnit = ref('')
  const deviceInfo = ref<Record<string, string>>({})
  // 设备压力单位一致性（计量设备与打压设备需同单位才能开始计量）。
  // 初始视为一致，避免首次进入未加载时误拦；连接/改单位/生成点位时会刷新。
  const unitConsistent = ref(true)

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
    precisionLevel: 0.0002,
    pressureMode: PressureMode.Single as PressureMode,
    controlMode: ControlMode.Auto as ControlMode
  })

  // 计量工作流相关
  const config = ref<MeasurementParamsPayload | null>(null)
  const points = ref<MeasurementPoint[]>([])
  const pointsEdited = ref(false)
  const pointsConfigKey = ref('')
  const currentPointIndex = ref(0)
  const alarmConfig = ref<MeasurementAlarmConfig>({
    enabled: true,
    enabledChannels: [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16],
    confirmOnAlarm: false,
    soundEnabled: true
  })
  const alarmPending = ref(false)
  const alarmData = ref<AlarmData | null>(null)
  const stabilityTimeoutPending = ref(false)

  // ── 计算属性 ──
  const runningStates: MeasurementState[] = ['pressurizing', 'stabilizing', 'collecting']
  const startableStates: MeasurementState[] = ['idle', 'completed', 'error', 'stopped']

  const isCollecting = computed(() => state.value === 'collecting')
  const isRunning = computed(() => runningStates.includes(state.value))
  const isStartable = computed(() => startableStates.includes(state.value))
  const isPaused = computed(() => state.value === 'paused')
  const isIdle = computed(() => state.value === 'idle')
  const totalRows = computed(() => rows.value.length)
  const deviceBound = computed(() => measureDeviceId.value !== '')
  const hasCompletedPoints = computed(() => points.value.some(p => p.status === 'completed'))

  // 阀门=校准模式是计量启动的必要条件，
  // 与标定模块共用同一规则。
  // 开关由 gate store 从后端 /api/v1/config/gates 拉取，避免前端硬编码。
  const valveReady = computed(() => valveStatus.value === 'calibration')
  const gatesStore = useGatesStore()
  const enforceValveCalibrationGate = computed(() => gatesStore.enforceValveCalibrationGate)
  // 开始计量必须满足：阀门=校准（若启用门禁）且设备压力单位一致。
  const canStart = computed(() => unitConsistent.value && (!enforceValveCalibrationGate.value || valveReady.value))

  // 主按钮：随会话状态自动切换文案、图标、色阶
  const primaryAction = computed<PrimaryAction>(() => {
    switch (state.value) {
      case 'idle':
      case 'ready':
      case 'stopped':
        return { key: 'start', label: '开始采集', icon: 'VideoPlay', variant: 'mint' }
      case 'pressurizing':
      case 'stabilizing':
      case 'collecting':
        return { key: 'pause', label: '暂停', icon: 'VideoPause', variant: 'slate' }
      case 'paused':
        return { key: 'resume', label: '继续采集', icon: 'VideoPlay', variant: 'mint' }
      case 'completed':
        return { key: 'export', label: '导出报告', icon: 'Download', variant: 'blue' }
      case 'error':
        return { key: 'retry', label: '重试', icon: 'Refresh', variant: 'amber' }
      default:
        return { key: 'start', label: '开始采集', icon: 'VideoPlay', variant: 'mint' }
    }
  })

  // 副按钮：仅在该状态确实可用时才出现
  const secondaryActions = computed<SecondaryAction[]>(() => {
    const out: SecondaryAction[] = []
    switch (state.value) {
      case 'pressurizing':
      case 'stabilizing':
      case 'collecting':
      case 'paused':
        out.push({ key: 'stop', label: '停止', variant: 'red', confirm: '确认终止采集？已采集数据将保留。' })
        break
      case 'completed':
        out.push({ key: 'restart', label: '重新采集', variant: 'slate', confirm: '将清空当前结果，重新开始？' })
        break
      case 'stopped':
        if (hasCompletedPoints.value) {
          out.push({ key: 'export', label: '导出报告', variant: 'blue' })
        }
        out.push({ key: 'reset', label: '清空数据', variant: 'slate', confirm: '将永久删除当前采集结果？' })
        break
      case 'error':
        out.push({ key: 'stop', label: '停止', variant: 'red', confirm: '确认终止当前会话？' })
        out.push({ key: 'view-error', label: '查看错误', variant: 'amber' })
        break
    }
    return out
  })

  // ── 设备绑定 ──

  const bindDevices = async (measureDevId: string, pressureDevId: string) => {
    await apiBindDevices(measureDevId, pressureDevId, 'measurement')
    measureDeviceId.value = measureDevId
    pressureDeviceId.value = pressureDevId
    await refreshUnitConsistency()
  }

  const bindMeasureDevice = async (measureDevId: string) => {
    await apiBindMeasureDevice(measureDevId, 'measurement')
    measureDeviceId.value = measureDevId
    await refreshUnitConsistency()
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

  // setValveStatus 切换阀门状态：
  // 成功时本地状态先按请求值更新，调用方应再通过 refreshValveStatus 校核硬件实际状态；
  // 失败时返回结构化错误（含 N09 等设备拒绝信息），便于 UI 弹出可读提示。
  const setValveStatus = async (status: string): Promise<ActionResult> => {
    try {
      await apiSetValveStatus(status)
      valveStatus.value = status
      return { ok: true }
    } catch (error) {
      const detail = error instanceof Error ? error.message : '设置阀门状态失败'
      console.error('设置阀门状态失败:', error)
      return { ok: false, error: 'VALVE_SET_FAILED', detail }
    }
  }

  const refreshMeasureUnit = async () => {
    for (let attempt = 1; attempt <= 3; attempt++) {
      try {
        measureUnit.value = await apiReadMeasureUnit()
        return
      } catch {
        if (attempt < 3) {
          await new Promise(r => setTimeout(r, 500))
        }
      }
    }
  }

  const setMeasureUnit = async (unit: string) => {
    await apiSetMeasureUnit(unit)
    measureUnit.value = unit
    await refreshUnitConsistency()
  }

  // 刷新设备压力单位一致性状态，用于开始计量的门禁。
  const refreshUnitConsistency = async () => {
    try {
      const check = await fetchUnitConsistency()
      unitConsistent.value = check?.consistent !== false
    } catch {
      // 拉取失败时不改变现有状态，避免因一次网络抖动误拦启动。
    }
  }

  const refreshDeviceInfo = async () => {
    try { deviceInfo.value = await apiReadDeviceInfo() }
    catch { /* 静默 */ }
  }

  const resetDevice = async () => {
    await apiResetDevice()
  }

  // ── 采集工作流 ──

  const start = async (selectedChannels: number[]): Promise<ActionResult> => {
    if (!deviceBound.value) {
      return { ok: false, error: 'DEVICE_NOT_BOUND', detail: '请先绑定计量设备' }
    }
    // 阀门=校准模式是启动的必要条件，前端先做一次轻量门禁，
    // 给用户即时反馈；后端会做权威校验，避免绕过。
    if (enforceValveCalibrationGate.value && !valveReady.value) {
      return { ok: false, error: 'VALVE_NOT_READY', detail: '请先将阀门切换到校准模式' }
    }
    try {
      await ensureDevicesBound()
      await syncPointsBeforeStart()
      channels.value = selectedChannels
      const newState = await startMeasurement(selectedChannels)
      state.value = newState as MeasurementState
      rows.value = []
      return { ok: true }
    } catch (error) {
      const detail = error instanceof Error ? error.message : String(error)
      return { ok: false, error: 'START_FAILED', detail }
    }
  }

  const manualStart = async (selectedChannels: number[]): Promise<ActionResult> => {
    if (!deviceBound.value) {
      return { ok: false, error: 'DEVICE_NOT_BOUND', detail: '请先绑定计量设备' }
    }
    if (enforceValveCalibrationGate.value && !valveReady.value) {
      return { ok: false, error: 'VALVE_NOT_READY', detail: '请先将阀门切换到校准模式' }
    }
    try {
      await ensureDevicesBound()
      await syncPointsBeforeStart()
      channels.value = selectedChannels
      currentPointIndex.value = 0
      const newState = await manualStartMeasurement(selectedChannels)
      state.value = newState as MeasurementState
      return { ok: true }
    } catch (error) {
      const detail = error instanceof Error ? error.message : String(error)
      return { ok: false, error: 'MANUAL_START_FAILED', detail }
    }
  }

  const pause = async (): Promise<ActionResult> => {
    try {
      const newState = await pauseMeasurement()
      state.value = newState as MeasurementState
      return { ok: true }
    } catch (error) {
      return { ok: false, error: 'PAUSE_FAILED', detail: '暂停采集失败' }
    }
  }

  const stop = async (): Promise<ActionResult> => {
    try {
      const newState = await stopMeasurement()
      state.value = newState as MeasurementState
      return { ok: true }
    } catch (error) {
      return { ok: false, error: 'STOP_FAILED', detail: '停止采集失败' }
    }
  }

  const refreshData = async () => {
    try {
      const resp = await fetchMeasurementData()
      // 后端已限制 s.rows 上限，此处再兜底截断：保留最近 MEASUREMENT_MAX_ROWS 行，
      // 确保任何异常返回都不会撑爆前端响应式 store 与表格 DOM。
      rows.value = resp.rows.length > MEASUREMENT_MAX_ROWS
        ? resp.rows.slice(-MEASUREMENT_MAX_ROWS)
        : resp.rows
    } catch { /* 静默 */ }
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

  const updateDevicePressure = (deviceId: string, pressure: number) => {
    const deviceStore = useMeasurementDeviceStore()
    deviceStore.updateDevicePressure(deviceId, pressure)
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
          pressureMode: cfg.pressureMode as PressureMode,
          controlMode: cfg.controlMode as ControlMode
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

  const buildMeasurementPayload = (p: typeof measurementParams.value, customPoints?: number[]): MeasurementParamsPayload => ({
    minPressure: p.minPressure,
    maxPressure: p.maxPressure,
    pointCount: p.pointCount,
    precision: p.precision,
    averageCount: p.averageCount,
    stableDurationMs: p.stableWaitS * 1000,
    precisionLevel: p.precisionLevel,
    pressureMode: p.pressureMode,
    controlMode: p.controlMode,
    ...(customPoints !== undefined ? { customPoints } : {})
  })

  const generatePoints = async (): Promise<ActionResult> => {
    try {
      const p = measurementParams.value
      const payload = buildMeasurementPayload(p)
      await saveMeasurementParamsConfig(payload)
      config.value = payload
      points.value = await generateMeasurementPoints()
      pointsEdited.value = false
      pointsConfigKey.value = measurementParamsKey(p)
      return { ok: true }
    } catch (error) {
      const detail = error instanceof Error ? error.message : String(error)
      return { ok: false, error: 'GENERATE_FAILED', detail }
    }
  }

  const syncPointsBeforeStart = async () => {
    const p = measurementParams.value
    const currentConfigKey = measurementParamsKey(p)
    const customPoints = pointsEdited.value && pointsConfigKey.value === currentConfigKey && points.value.length > 0
      ? points.value.map(point => point.targetPressure)
      : undefined
    const payload = buildMeasurementPayload(p, customPoints)
    await saveMeasurementParamsConfig(payload)
    config.value = payload
    points.value = await generateMeasurementPoints()
    pointsEdited.value = false
    pointsConfigKey.value = currentConfigKey
  }

  const ensureDevicesBound = async () => {
    if (!measureDeviceId.value) return
    if (pressureDeviceId.value) {
      await apiBindDevices(measureDeviceId.value, pressureDeviceId.value)
      return
    }
    await apiBindMeasureDevice(measureDeviceId.value)
  }

  const measurementParamsKey = (p: typeof measurementParams.value): string => JSON.stringify({
    minPressure: p.minPressure,
    maxPressure: p.maxPressure,
    pointCount: p.pointCount,
    precision: p.precision,
    pressureMode: p.pressureMode
  })

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

  const resolveAlarm = async (decision: AlarmDecision) => {
    await resolveMeasurementAlarm(decision)
    alarmPending.value = false
  }

  // ── 按点采集 ──

  const startPoint = (index: number) => {
    currentPointIndex.value = index
  }

  const completePoint = () => {
    currentPointIndex.value++
  }

  const resetCollection = () => {
    currentPointIndex.value = 0
    rows.value = []
    points.value = []
    pointsEdited.value = false
    pointsConfigKey.value = ''
    state.value = 'idle'
  }

  const updatePointTarget = (pointId: string, targetPressure: number) => {
    const pt = points.value.find(p => p.id === pointId)
    if (pt) {
      pt.targetPressure = targetPressure
      pointsEdited.value = true
    }
  }

  const autoCollect = async (): Promise<ActionResult> => {
    try {
      const newState = await autoCollectMeasurement()
      state.value = newState as MeasurementState
      return { ok: true }
    } catch (error) {
      const detail = error instanceof Error ? error.message : String(error)
      return { ok: false, error: 'AUTO_COLLECT_FAILED', detail }
    }
  }

  const manualPressurize = async (pointIndex: number): Promise<ActionResult> => {
    try {
      const newState = await manualPressurizeMeasurement(pointIndex)
      state.value = newState as MeasurementState
      return { ok: true }
    } catch (error) {
      const detail = error instanceof Error ? error.message : String(error)
      return { ok: false, error: 'MANUAL_PRESSURIZE_FAILED', detail }
    }
  }

  const manualCollect = async (pointIndex: number): Promise<ActionResult> => {
    try {
      const newState = await manualCollectMeasurement(pointIndex)
      state.value = newState as MeasurementState
      return { ok: true }
    } catch (error) {
      const detail = error instanceof Error ? error.message : String(error)
      return { ok: false, error: 'MANUAL_COLLECT_FAILED', detail }
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
    unitConsistent,
    stabilityState,
    stabilityTimeoutPending,
    // 计算属性
    isCollecting,
    isRunning,
    isStartable,
    isPaused,
    isIdle,
    totalRows,
    deviceBound,
    hasCompletedPoints,
    valveReady,
    canStart,
    primaryAction,
    secondaryActions,
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
    refreshUnitConsistency,
    resetDevice,
    // 采集工作流
    start,
    manualStart,
    pause,
    stop,
    refreshData,
    // 计量工作流
    config,
    points,
    currentPointIndex,
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
    startPoint,
    completePoint,
    resetCollection,
    updatePointTarget,
    autoCollect,
    manualPressurize,
    manualCollect,
    fetchCurrentState,
    syncState,
    updateDevicePressure
  }
})
