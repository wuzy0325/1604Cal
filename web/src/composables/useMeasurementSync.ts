import { onMounted, onUnmounted } from 'vue'
import { useEventHub } from '@/composables/useEventHub'
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

  const { subscribe, registerPoll } = useEventHub()
  const unsubs: (() => void)[] = []
  const unregPolls: (() => void)[] = []

  onMounted(async () => {
    await Promise.all([
      store.loadAlarmConfig(),
      store.fetchCurrentState()
    ])

    unsubs.push(subscribe(EVENT_MEASUREMENT_STATE_CHANGED, (payload) => {
      const newState = (payload.data as { state: MeasurementState }).state
      store.syncState(newState)
      // 进入打压状态时清除旧报警标记，避免上个点的报警残留到当前点
      if (newState === 'pressurizing') {
        store.alarmData = null
      }
    }))

    unsubs.push(subscribe(EVENT_MEASUREMENT_DATA_UPDATED, (payload) => {
      const data = payload.data as { timestamp: string; channels: Record<string, number> }
      store.rows.push({ timestamp: data.timestamp, channels: data.channels })
    }))

    unsubs.push(subscribe(EVENT_MEASUREMENT_STABILITY_UPDATE, (payload) => {
      store.stabilityState = payload.data as StabilityUpdate
    }))

    unsubs.push(subscribe(EVENT_MEASUREMENT_ALARM_TRIGGERED, (payload) => {
      store.alarmPending = true
      store.alarmData = payload.data as AlarmData
    }))

    unsubs.push(subscribe(EVENT_MEASUREMENT_STABILITY_TIMEOUT, () => {
      store.stabilityTimeoutPending = true
    }))

    unsubs.push(subscribe(EVENT_MEASUREMENT_ALARM_RESOLVED, () => {
      store.alarmPending = false
      store.alarmData = null
    }))

    unsubs.push(subscribe(EVENT_MEASUREMENT_POINT_STATUS, (payload) => {
      const updated = payload.data as MeasurementPoint
      const idx = store.points.findIndex(p => p.id === updated.id)
      if (idx >= 0) store.points[idx] = updated
    }))

    unsubs.push(subscribe(EVENT_MEASUREMENT_DATA_COLLECTED, (payload) => {
      const collected = payload.data as { pointIndex: number; channels: number[]; data: number[] }
      const ptIdx = store.points.findIndex(p => p.index === collected.pointIndex)
      if (ptIdx >= 0) {
        store.points[ptIdx] = { ...store.points[ptIdx], collectedData: collected.data, status: 'completed' }
      }
    }))

    unsubs.push(subscribe(EVENT_MULTIPRESS_PRESSURE_UPDATE, (payload) => {
      const data = payload.data as Record<string, unknown>
      const deviceId = data?.deviceId as string | undefined
      const pressure = data?.currentPressure as number | undefined
      if (deviceId && typeof pressure === 'number') {
        store.updateDevicePressure(deviceId, pressure)
        deviceStore.updateDevicePressure(deviceId, pressure)
      }
    }))

    unregPolls.push(registerPoll('measurement-running', async () => {
      if (store.isRunning) {
        await Promise.all([
          store.refreshPressure(),
          store.refreshStability(),
          store.refreshMeasureData()
        ])
      }
    }, 2000))

    unregPolls.push(registerPoll('measurement-pressure', async () => {
      try {
        const states = await multipressListDevices()
        for (const s of states) {
          deviceStore.updateDevicePressure(s.deviceId, s.currentPressure)
        }
      } catch {
        // 静默失败
      }
    }, 1000))
  })

  onUnmounted(() => {
    for (const unsub of unsubs) unsub()
    for (const unreg of unregPolls) unreg()
  })
}
