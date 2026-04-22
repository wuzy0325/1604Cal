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
  getMeasurementExportUrl
} from '@/api/measurement'
import { createEventStream } from '@/api/client'
import type { StreamEventPayload } from '@/types/api'
import type { MeasurementState, CollectedRow } from './types'

export type { MeasurementState, CollectedRow }

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

  // SSE
  let eventSource: EventSource | null = null

  // ── 计算属性 ──
  const isCollecting = computed(() => state.value === 'collecting')
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

  return {
    // 状态
    state,
    rows,
    channels,
    measureDeviceId,
    pressureDeviceId,
    currentPressure,
    isStable,
    channelData,
    valveStatus,
    measureUnit,
    deviceInfo,
    // 计算属性
    isCollecting,
    isPaused,
    isIdle,
    totalRows,
    exportUrl,
    deviceBound,
    // 设备绑定
    bindDevices,
    bindMeasureDevice,
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
    // SSE
    setupSSE,
    teardownSSE,
    fetchCurrentState,
    syncState
  }
})
