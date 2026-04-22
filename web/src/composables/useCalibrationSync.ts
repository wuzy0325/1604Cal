import { onMounted, onUnmounted, ref } from 'vue'
import { createEventStream } from '@/api/client'
import { connectDevice } from '@/api/device'
import { bindMeasureDevice } from '@/api/session'
import type { SessionState } from '@/types/calibration'
import type { StreamEventPayload } from '@/types/api'
import { useCalibrationStore } from '@/stores/calibration'
import { useDeviceInventoryStore } from '@/stores/device/inventoryStore'

// 稳定性 SSE 事件数据结构
export interface StabilityEventData {
  isStable: boolean
  isInRange: boolean
  currentValue: number
  targetValue: number
  deviation: number
  stableDurationMs: number
  requiredDurationMs: number
  progress: number
}

// 报警 SSE 事件数据结构
export interface AlarmEventData {
  pointIndex: number
  targetPressure: number
  overLimitChannels: number[]
  maxDeviation: number
  channelDetails: Record<string, number>
}

/**
 * Composable that manages SSE event stream and polling for calibration view.
 * Automatically starts on mount and cleans up on unmount.
 */
export function useCalibrationSync() {
  const calibrationStore = useCalibrationStore()
  const deviceStore = useDeviceInventoryStore()

  let eventSource: EventSource | null = null
  let pollTimer: ReturnType<typeof setInterval> | null = null
  let deviceRefreshTimer: ReturnType<typeof setInterval> | null = null
  let bindingInProgress = false
  let boundMeasureId = ''
  let lastRepairAttemptAt = 0

  // 稳定性状态
  const stabilityStatus = ref<StabilityEventData | null>(null)
  // 报警事件状态
  const alarmEvent = ref<AlarmEventData | null>(null)

  // 绑定计量设备并刷新阀门/单位信息。
  // 兼容“设备状态显示已连接但会话驱动未就绪”的场景，必要时静默触发一次重连修复。
  async function bindConnectedMeasureDevice() {
    if (bindingInProgress) return

    const connectedMeasure = deviceStore.measureDevices.find(d => d.status === 'connected')
    if (!connectedMeasure) {
      boundMeasureId = ''
      return
    }

    // 若当前设备已完成绑定且阀门/单位信息已可读，跳过重复绑定。
    if (
      boundMeasureId === connectedMeasure.id &&
      calibrationStore.valveStatus !== '' &&
      calibrationStore.measureUnit !== ''
    ) {
      return
    }

    bindingInProgress = true
    try {
      try {
        await bindMeasureDevice(connectedMeasure.id)
      } catch {
        // 绑定失败由后续读取结果决定是否进入修复流程
      }

      let loaded = await calibrationStore.refreshDeviceInfo({ retries: 3, retryDelayMs: 500 })
      if (!loaded) {
        const now = Date.now()
        // 避免高频重连，最多每 15 秒尝试一次静默修复。
        if (now - lastRepairAttemptAt >= 15000) {
          lastRepairAttemptAt = now
          try {
            await connectDevice(connectedMeasure.id)
          } catch {
            // 静默修复失败，不弹窗；等待下一轮刷新重试
          }

          await deviceStore.loadDevices(true)
          try {
            await bindMeasureDevice(connectedMeasure.id)
          } catch {
            // 忽略，交给后续 refreshDeviceInfo 判断
          }

          loaded = await calibrationStore.refreshDeviceInfo({ retries: 3, retryDelayMs: 500 })
        }
      }

      if (loaded) {
        boundMeasureId = connectedMeasure.id
      }
    } finally {
      bindingInProgress = false
    }
  }

  function setupSSE() {
    eventSource = createEventStream((payload: StreamEventPayload) => {
      if (payload.type === 'session.state.changed') {
        const data = payload.data as { state: SessionState }
        if (data?.state) {
          calibrationStore.syncSessionState(data.state)
        }
      }
      if (payload.type === 'device.status.changed') {
        void deviceStore.loadDevices(true).then(bindConnectedMeasureDevice)
      }
      // 稳定性 SSE 事件
      if (payload.type?.startsWith('calibration.stability.')) {
        stabilityStatus.value = payload.data as StabilityEventData
      }
      // 报警事件
      if (payload.type === 'alarm.triggered') {
        alarmEvent.value = payload.data as AlarmEventData
      }
    })
  }

  function startPolling() {
    if (pollTimer) return
    pollTimer = setInterval(async () => {
      if (calibrationStore.isRunning) {
        await Promise.all([
          calibrationStore.refreshPressure(),
          calibrationStore.refreshStability(),
          calibrationStore.refreshMeasureData()
        ])
      }
    }, 2000)
  }

  function startDeviceRefresh() {
    if (deviceRefreshTimer) return
    deviceRefreshTimer = setInterval(() => {
      void deviceStore.loadDevices(true).then(bindConnectedMeasureDevice)
    }, 5000)
  }

  function stopPolling() {
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
    if (deviceRefreshTimer) {
      clearInterval(deviceRefreshTimer)
      deviceRefreshTimer = null
    }
  }

  onMounted(async () => {
    await deviceStore.loadDevices()
    await calibrationStore.fetchCurrentSessionState()

    await bindConnectedMeasureDevice()

    setupSSE()
    startPolling()
    startDeviceRefresh()
  })

  onUnmounted(() => {
    if (eventSource) {
      eventSource.close()
      eventSource = null
    }
    stopPolling()
  })

  return { stabilityStatus, alarmEvent }
}
