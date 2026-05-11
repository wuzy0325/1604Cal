import { onMounted, onUnmounted } from 'vue'
import { createEventStream } from '@/api/client'
import type { StreamEventPayload } from '@/types/api'
import { useMeasurementStore } from '@/stores/measurement'
import { useDeviceInventoryStore } from '@/stores/device/inventoryStore'
import type { MeasurementState, StabilityUpdate, AlarmData } from '@/stores/measurement/types'
import type { MeasurementPoint } from '@/api/measurement'
import {
  EVENT_MEASUREMENT_STATE_CHANGED,
  EVENT_MEASUREMENT_DATA_UPDATED,
  EVENT_MEASUREMENT_STABILITY_UPDATE,
  EVENT_MEASUREMENT_STABILITY_TIMEOUT,
  EVENT_MEASUREMENT_ALARM_TRIGGERED,
  EVENT_MEASUREMENT_ALARM_RESOLVED,
  EVENT_MEASUREMENT_POINT_STATUS,
  EVENT_MEASUREMENT_DATA_COLLECTED,
  EVENT_MULTIPRESS_PRESSURE_UPDATE,
} from '@/shared/events'
import { multipressListDevices } from '@/api/multipress'

/**
 * Composable that manages SSE event stream for measurement view.
 * Mirrors the calibration module's useCalibrationSync pattern:
 * SSE lifecycle is owned by the composable, not the store.
 */
export function useMeasurementSync() {
  const store = useMeasurementStore()
  const deviceStore = useDeviceInventoryStore()

  let eventSource: EventSource | null = null

  function setupSSE() {
    if (eventSource) return
    eventSource = createEventStream((payload: StreamEventPayload) => {
      switch (payload.type) {
        case EVENT_MEASUREMENT_STATE_CHANGED:
          store.syncState((payload.data as { state: MeasurementState }).state)
          break
        case EVENT_MEASUREMENT_DATA_UPDATED: {
          const data = payload.data as { timestamp: string; channels: Record<string, number> }
          store.rows.push({ timestamp: data.timestamp, channels: data.channels })
          break
        }
        case EVENT_MEASUREMENT_STABILITY_UPDATE: {
          const status = payload.data as StabilityUpdate
          store.stabilityState = status
          break
        }
        case EVENT_MEASUREMENT_ALARM_TRIGGERED:
          store.alarmPending = true
          store.alarmData = payload.data as AlarmData
          break
        case EVENT_MEASUREMENT_STABILITY_TIMEOUT:
          store.stabilityTimeoutPending = true
          break
        case EVENT_MEASUREMENT_ALARM_RESOLVED:
          store.alarmPending = false
          // 保留 alarmData 不清理，红色标记持续到新数据覆盖
          break
        case EVENT_MEASUREMENT_POINT_STATUS: {
          const updated = payload.data as MeasurementPoint
          const idx = store.points.findIndex(p => p.id === updated.id)
          if (idx >= 0) store.points[idx] = updated
          break
        }
        case EVENT_MEASUREMENT_DATA_COLLECTED: {
          const collected = payload.data as { pointIndex: number; channels: number[]; data: number[] }
          const ptIdx = store.points.findIndex(p => p.index === collected.pointIndex)
          if (ptIdx >= 0) {
            store.points[ptIdx] = { ...store.points[ptIdx], collectedData: collected.data, status: 'completed' }
          }
          // 新采集数据到达，清除旧报警标记；若仍超限后续 alarm.triggered 会重新设置
          store.alarmData = null
          break
        }
        case EVENT_MULTIPRESS_PRESSURE_UPDATE: {
          const data = payload.data as Record<string, unknown>
          const deviceId = data?.deviceId as string | undefined
          const pressure = data?.currentPressure as number | undefined
          if (deviceId && typeof pressure === 'number') {
            store.updateDevicePressure(deviceId, pressure)
            deviceStore.updateDevicePressure(deviceId, pressure)
          }
          break
        }
      }
    })
  }

  function teardownSSE() {
    if (eventSource) {
      eventSource.close()
      eventSource = null
    }
  }

  let pollTimer: ReturnType<typeof setInterval> | null = null
  let pressureRefreshTimer: ReturnType<typeof setInterval> | null = null

  function startPolling() {
    if (pollTimer) return
    pollTimer = setInterval(async () => {
      if (store.isRunning) {
        await Promise.all([
          store.refreshPressure(),
          store.refreshStability(),
          store.refreshMeasureData()
        ])
      }
    }, 2000)
  }

  function startPressureRefresh() {
    if (pressureRefreshTimer) return
    pressureRefreshTimer = setInterval(async () => {
      try {
        const states = await multipressListDevices()
        for (const s of states) {
          deviceStore.updateDevicePressure(s.deviceId, s.currentPressure)
        }
      } catch {
        // 静默失败
      }
    }, 1000)
  }

  function stopPolling() {
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
    if (pressureRefreshTimer) {
      clearInterval(pressureRefreshTimer)
      pressureRefreshTimer = null
    }
  }

  onMounted(async () => {
    await Promise.all([
      store.loadAlarmConfig(),
      store.fetchCurrentState()
    ])
    setupSSE()
    startPolling()
    startPressureRefresh()
  })

  onUnmounted(() => {
    teardownSSE()
    stopPolling()
  })
}
