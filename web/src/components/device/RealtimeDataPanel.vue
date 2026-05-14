<template>
  <section class="realtime-data-panel">
    <header class="panel-header">
      <div class="header-title">
        <el-icon class="panel-icon">
          <DataLine />
        </el-icon>
        <h2>实时数据监控</h2>
      </div>
      <div class="header-actions">
        <div class="device-status-group">
          <span
            class="device-status-badge"
            :class="measureDeviceStatusClass"
          >
            计量: {{ measureDeviceStatusText }}
          </span>
          <span
            class="device-status-badge"
            :class="pressureDeviceStatusClass"
          >
            打压: {{ pressureDeviceStatusText }}
          </span>
        </div>
        <span class="update-time">{{ lastUpdateTime }}</span>
        <button
          type="button"
          class="icon-btn"
          title="重新连接"
          @click="reconnect"
        >
          <el-icon><Refresh /></el-icon>
        </button>
      </div>
    </header>

    <div class="metrics-bar">
      <div class="metric-item pressure-current">
        <span class="metric-label">当前压力</span>
        <span class="metric-value">{{ formatPressure(currentPressure) }}</span>
        <span class="metric-unit">kPa</span>
        <div
          class="mini-bar"
          role="progressbar"
          :aria-valuenow="currentPressure"
          aria-valuemin="0"
          :aria-valuemax="targetPressureForMath"
        >
          <div
            class="mini-bar-fill"
            :style="{ width: pressurePercentage + '%' }"
            :class="pressureBarClass"
          />
        </div>
      </div>

      <div class="metric-separator" />

      <div class="metric-item pressure-target">
        <span class="metric-label">目标压力</span>
        <span class="metric-value">{{ formatPressure(targetPressure) }}</span>
        <span class="metric-unit">kPa</span>
        <span
          class="target-diff"
          :class="targetDiffClass"
        >
          {{ targetDiffText }}
        </span>
      </div>

      <div class="metric-separator" />

      <div class="metric-item stability">
        <span class="metric-label">稳定状态</span>
        <span
          class="status-indicator"
          :class="stabilityClass"
        />
        <span
          class="status-text"
          :class="stabilityClass"
        >{{ stabilityText }}</span>
        <span class="stability-meta">{{ stableDuration }}s / ±{{ stabilityThreshold }}kPa</span>
      </div>
    </div>

    <!-- 通道数据表格 -->
    <div class="channels-section">
      <div class="section-header">
        <div class="section-title">
          <el-icon><Grid /></el-icon>
          <h3>通道读数</h3>
        </div>
        <span class="channel-count">{{ activeChannels }}/{{ totalChannels }} 通道活跃</span>
      </div>
      <div class="channels-grid">
        <div
          v-for="(channel, index) in channelData"
          :key="index"
          class="channel-item"
          :class="{ 'channel-active': channel.isActive }"
        >
          <div class="channel-header">
            <span class="channel-name">CH{{ index + 1 }}</span>
            <span
              class="channel-status"
              :class="channel.status"
            >
              {{ channelStatusText(channel.status) }}
            </span>
          </div>
          <div class="channel-value">
            {{ formatChannelValue(channel.value) }}
          </div>
        </div>
      </div>
    </div>

    <!-- 采集进度 -->
    <div
      v-if="showProgress"
      class="progress-section"
    >
      <div class="progress-header">
        <div class="progress-title">
          <el-icon><Histogram /></el-icon>
          <span>采集进度</span>
        </div>
        <span class="progress-percent">{{ progressPercent }}%</span>
      </div>
      <div
        class="progress-bar"
        role="progressbar"
        :aria-valuenow="completedPoints"
        aria-valuemin="0"
        :aria-valuemax="totalPoints"
      >
        <div
          class="progress-fill"
          :style="{ width: progressPercent + '%' }"
        />
      </div>
      <div class="progress-meta">
        <span>已完成: {{ completedPoints }}/{{ totalPoints }} 点</span>
        <span>预计剩余: {{ estimatedTime }}</span>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import {
  DataLine,
  Refresh,
  Grid,
  Histogram
} from '@element-plus/icons-vue'
import { createEventStream } from "@/api/client"
import { fetchDevices } from "@/api/device"
import {
  readPressure,
  readStability,
  readMeasureData
} from "@/api/session"
import type { StreamEventPayload } from "@/types/api"
import { useDeviceStore } from '@/stores/deviceStore'

interface ChannelInfo {
  value: number | null
  status: 'ok' | 'warning' | 'error' | 'idle'
  isActive: boolean
}

const props = withDefaults(defineProps<{
  targetPressure?: number | null
  completedPoints?: number
  totalPoints?: number
}>(), {
  targetPressure: null,
  completedPoints: 0,
  totalPoints: 0
})

// 响应式数据
const currentPressure = ref(0)
const eventTargetPressure = ref<number | null>(null)
const isStable = ref(false)
const stableDuration = ref(0)
const stabilityThreshold = ref(0.5)
const lastUpdateTime = ref('--:--:--')
const isConnected = ref(false)
const measureDeviceStatus = ref<'connected' | 'disconnected' | 'not_selected'>('not_selected')
const pressureDeviceStatus = ref<'connected' | 'disconnected' | 'not_selected'>('not_selected')

const channelData = ref<ChannelInfo[]>(
  Array.from({ length: 16 }, () => ({
    value: null,
    status: 'idle',
    isActive: false
  }))
)

const deviceStore = useDeviceStore()

const totalPoints = computed(() => Math.max(0, Math.round(props.totalPoints)))

const completedPoints = computed(() => {
  const raw = Math.max(0, Math.round(props.completedPoints))
  if (totalPoints.value === 0) {
    return 0
  }

  return Math.min(totalPoints.value, raw)
})

const showProgress = computed(() => totalPoints.value > 0)

const estimatedTime = computed(() => {
  const remaining = totalPoints.value - completedPoints.value
  if (remaining <= 0) {
    return '00:00'
  }

  return '--'
})

// 计算属性
const measureDeviceStatusText = computed(() => {
  switch (measureDeviceStatus.value) {
    case 'connected': return '已连接'
    case 'disconnected': return '未连接'
    default: return '未选择'
  }
})

const pressureDeviceStatusText = computed(() => {
  switch (pressureDeviceStatus.value) {
    case 'connected': return '已连接'
    case 'disconnected': return '未连接'
    default: return '未选择'
  }
})

const measureDeviceStatusClass = computed(() => ({
  'status-connected': measureDeviceStatus.value === 'connected',
  'status-disconnected': measureDeviceStatus.value === 'disconnected',
  'status-not-selected': measureDeviceStatus.value === 'not_selected'
}))

const pressureDeviceStatusClass = computed(() => ({
  'status-connected': pressureDeviceStatus.value === 'connected',
  'status-disconnected': pressureDeviceStatus.value === 'disconnected',
  'status-not-selected': pressureDeviceStatus.value === 'not_selected'
}))

const targetPressure = computed<number | null>(() => {
  if (typeof props.targetPressure === 'number' && Number.isFinite(props.targetPressure)) {
    return props.targetPressure
  }

  if (typeof eventTargetPressure.value === 'number' && Number.isFinite(eventTargetPressure.value)) {
    return eventTargetPressure.value
  }

  return null
})

const targetPressureForMath = computed(() => {
  if (typeof targetPressure.value === 'number' && targetPressure.value > 0) {
    return targetPressure.value
  }

  return 0
})

const pressurePercentage = computed(() => {
  if (targetPressureForMath.value <= 0) return 0
  const percent = (currentPressure.value / targetPressureForMath.value) * 100
  return Math.min(Math.max(percent, 0), 100)
})

const pressureBarClass = computed(() => {
  if (targetPressure.value === null) return 'bar-normal'
  if (isStable.value) return 'bar-stable'
  if (Math.abs(currentPressure.value - targetPressure.value) <= stabilityThreshold.value) {
    return 'bar-approaching'
  }
  return 'bar-normal'
})

const stabilityClass = computed(() => ({
  'indicator-stable': isStable.value,
  'indicator-unstable': !isStable.value
}))

const stabilityText = computed(() =>
  isStable.value ? '已稳定' : '未稳定'
)

const targetDiff = computed<number | null>(() => {
  if (targetPressure.value === null) {
    return null
  }

  return currentPressure.value - targetPressure.value
})

const targetDiffClass = computed(() => {
  if (targetDiff.value === null) return 'diff-unknown'
  const diff = Math.abs(targetDiff.value)
  if (diff <= stabilityThreshold.value) return 'diff-ok'
  if (diff <= stabilityThreshold.value * 3) return 'diff-warning'
  return 'diff-error'
})

const targetDiffText = computed(() => {
  if (targetDiff.value === null) {
    return '--'
  }

  const diff = targetDiff.value
  const sign = diff > 0 ? '+' : ''
  return `${sign}${formatPressure(diff)} kPa`
})

const activeChannels = computed(() =>
  channelData.value.filter(ch => ch.isActive).length
)

const totalChannels = computed(() => 16)

const progressPercent = computed(() => {
  if (totalPoints.value <= 0) {
    return 0
  }

  return Math.round((completedPoints.value / totalPoints.value) * 100)
})

// 方法
function formatPressure(value: number | null): string {
  if (value === null) return '--.--'
  return value.toFixed(2)
}

function formatChannelValue(value: number | null): string {
  if (value === null) return '---'
  return value.toFixed(3)
}

function channelStatusText(status: string): string {
  const map: Record<string, string> = {
    ok: '正常',
    warning: '警告',
    error: '异常',
    idle: '空闲'
  }
  return map[status] || status
}

// SSE 事件订阅
let eventSource: EventSource | null = null
let pollInterval: ReturnType<typeof setInterval> | null = null

function setupSSE() {
  eventSource = createEventStream({
    onEvent: (payload: StreamEventPayload) => {
      const now = new Date()
      lastUpdateTime.value = now.toLocaleTimeString('zh-CN')

      if (payload.type === 'device.status.changed') {
        // 设备状态变化时，重新同步连接状态
        void syncConnectionFromDevices()
      }

      if (payload.type === 'pressure.applied') {
        const data = payload.data as { actualPressure?: number; targetPressure?: number }
        if (data?.actualPressure !== undefined) {
          currentPressure.value = data.actualPressure
        }
        if (data?.targetPressure !== undefined) {
          eventTargetPressure.value = data.targetPressure
        }
      }

      if (payload.type === 'data.collected') {
        const data = payload.data as { data?: number[]; channels?: number[] }
        if (data?.data && data?.channels) {
          updateChannelData(data.data, data.channels)
        }
      }
    },
    onError: (error) => {
      console.warn('[RealtimeDataPanel] SSE 连接断开:', error)
    }
  })
}

function updateChannelData(values: number[], channels: number[]) {
  const newChannelData: ChannelInfo[] = Array.from({ length: 16 }, () => ({
    value: null,
    status: 'idle' as ChannelInfo['status'],
    isActive: false
  }))

  channels.forEach((ch, idx) => {
    if (ch >= 1 && ch <= 16 && idx < values.length) {
      const chIdx = ch - 1
      const currentValue = values[idx]
      const compareTarget = targetPressure.value ?? currentValue
      newChannelData[chIdx] = {
        value: currentValue,
        status: Math.abs(currentValue - compareTarget) > stabilityThreshold.value * 3 ? 'warning' : 'ok',
        isActive: true
      }
    }
  })

  channelData.value = newChannelData
}

// 轮询实时数据（作为SSE的补充）
async function pollRealtimeData() {
  try {
    const pressure = await readPressure()
    currentPressure.value = pressure
    isConnected.value = true
  } catch {
    // 设备未连接或不可用
  }

  try {
    const stable = await readStability()
    isStable.value = stable
    if (stable) {
      stableDuration.value += 1
    } else {
      stableDuration.value = 0
    }
  } catch {
    // 忽略
  }

  try {
    const data = await readMeasureData()
    if (data.length > 0) {
      const channels = Array.from({ length: Math.min(data.length, 16) }, (_, i) => i + 1)
      updateChannelData(data, channels)
    }
  } catch {
    // 忽略
  }
}

function reconnect() {
  if (eventSource) {
    eventSource.close()
  }
  setupSSE()
  syncConnectionFromDevices()
}

/** 从后端设备列表同步连接状态，解决进入页面时设备已连接但未触发 SSE 的问题。 */
async function syncConnectionFromDevices() {
  try {
    const devices = await fetchDevices()
    const selection = deviceStore.selectionByModule('measurement')
    const pressureId = selection.pressureDeviceId
    const measureId = selection.measureDeviceId

    // 检查两个设备是否都已连接
    const pressureDevice = devices.find(d => d.id === pressureId)
    const measureDevice = devices.find(d => d.id === measureId)

    // 更新各设备状态
    measureDeviceStatus.value = measureId
      ? (measureDevice?.status === 'connected' ? 'connected' : 'disconnected')
      : 'not_selected'
    pressureDeviceStatus.value = pressureId
      ? (pressureDevice?.status === 'connected' ? 'connected' : 'disconnected')
      : 'not_selected'

    // 只有当两个设备都已选择且都已连接时，才认为已连接
    const pressureConnected = pressureId && pressureDevice?.status === 'connected'
    const measureConnected = measureId && measureDevice?.status === 'connected'

    isConnected.value = !!(pressureConnected && measureConnected)
  } catch {
    // 查询失败时保持当前状态
  }
}

// 生命周期
onMounted(() => {
  setupSSE()
  syncConnectionFromDevices()
  // 每2秒轮询一次实时数据
  pollInterval = setInterval(pollRealtimeData, 2000)
  // 立即执行一次
  pollRealtimeData()
})

onUnmounted(() => {
  if (eventSource) {
    eventSource.close()
    eventSource = null
  }
  if (pollInterval) {
    clearInterval(pollInterval)
    pollInterval = null
  }
})
</script>

<style scoped lang="scss">
.realtime-data-panel {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  padding: var(--spacing-sm);
  height: 100%;
  min-height: 0;
  display: grid;
  grid-template-rows: auto auto minmax(0, 1fr) auto;
  gap: var(--spacing-sm);
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0;
  padding-bottom: var(--spacing-xs);
  border-bottom: 1px solid var(--border-color);
}

.header-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

.panel-icon {
  font-size: 20px;
  color: var(--accent-primary);
}

.header-title h2 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.connection-status {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);
  padding: 3px 8px;
  border-radius: 3px;
  font-size: 11px;
  font-weight: 600;

  .el-icon {
    font-size: 11px;
  }
}

.status-connected {
  background: var(--status-success-bg);
  color: var(--status-success);
}

.status-disconnected {
  background: var(--status-error-bg);
  color: var(--status-error);
}

.status-not-selected {
  background: var(--bg-tertiary);
  color: var(--text-muted);
}

.device-status-group {
  display: flex;
  gap: var(--spacing-xs);
}

.device-status-badge {
  display: inline-flex;
  align-items: center;
  padding: 3px 8px;
  border-radius: 3px;
  font-size: 11px;
  font-weight: 600;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

.update-time {
  font-size: 12px;
  color: var(--text-muted);
  font-family: Consolas, monospace;
}

.icon-btn {
  width: 28px;
  height: 28px;
  border: 1px solid var(--border-color-strong);
  border-radius: 3px;
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;

  .el-icon {
    font-size: 14px;
  }

  &:hover {
    background: var(--border-color);
    border-color: var(--accent-primary);
    color: var(--accent-primary);
  }
}

/* ===== 紧凑指标栏 ===== */
.metrics-bar {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 3px;
  padding: var(--spacing-xs) var(--spacing-md);
}

.metric-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  white-space: nowrap;
}

.metric-label {
  font-size: 11px;
  font-weight: 500;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.metric-value {
  font-family: Consolas, monospace;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  line-height: 1;
}

.metric-unit {
  font-size: 11px;
  color: var(--text-muted);
  font-weight: 500;
}

.metric-separator {
  width: 1px;
  height: 20px;
  background: var(--border-color-strong);
  flex-shrink: 0;
}

/* 迷你压力条 */
.mini-bar {
  width: 60px;
  height: 4px;
  background: var(--bg-primary);
  border-radius: 2px;
  overflow: hidden;
  margin-left: var(--spacing-xs);
}

.mini-bar-fill {
  height: 100%;
  border-radius: 2px;
  transition: width 0.3s ease, background-color 0.3s ease;
}

.bar-normal { background: var(--status-info); }
.bar-approaching { background: var(--status-warning); }
.bar-stable { background: var(--status-success); }

/* 目标差值 */
.target-diff {
  font-size: 11px;
  font-weight: 500;
  padding: 1px 5px;
  border-radius: 2px;
  font-family: Consolas, monospace;
}

.diff-ok { background: var(--status-success-bg); color: var(--status-success); }
.diff-warning { background: var(--status-warning-bg); color: var(--status-warning); }
.diff-error { background: var(--status-error-bg); color: var(--status-error); }
.diff-unknown { background: var(--bg-quaternary); color: var(--text-muted); }

/* 稳定状态 */
.status-indicator {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  transition: background-color 0.3s ease;
}

.indicator-stable {
  background: var(--status-success);
  box-shadow: 0 0 0 2px var(--status-success-bg);
}

.indicator-unstable {
  background: var(--status-warning);
}

.status-text {
  font-size: 12px;
  font-weight: 600;
}

.status-text.indicator-stable { color: var(--status-success); }
.status-text.indicator-unstable { color: var(--status-warning); }

.stability-meta {
  font-size: 10px;
  color: var(--text-muted);
  font-family: Consolas, monospace;
}

.channels-section {
  margin-bottom: 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-sm);
}

.section-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);

  .el-icon {
    font-size: 16px;
    color: var(--accent-primary);
  }
}

.section-header h3 {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.channel-count {
  font-size: 11px;
  color: var(--text-secondary);
  background: var(--bg-tertiary);
  padding: 2px 6px;
  border-radius: 3px;
}

.channels-grid {
  display: grid;
  grid-template-columns: repeat(8, 1fr);
  gap: var(--spacing-xs);
  min-height: 0;
  overflow: auto;
  align-content: start;
  padding-right: 2px;
}

.channel-item {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 2px;
  padding: 4px;
  text-align: center;
}

.channel-item.channel-active {
  border-color: var(--accent-primary);
  background: rgba(255, 215, 0, 0.06);
}

.channel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2px;
}

.channel-name {
  font-size: 10px;
  font-weight: 600;
  color: var(--text-muted);
}

.channel-status {
  font-size: 9px;
  padding: 1px 3px;
  border-radius: 2px;
  font-weight: 500;
}

.channel-status.ok {
  background: var(--status-success-bg);
  color: var(--status-success);
}

.channel-status.warning {
  background: var(--status-warning-bg);
  color: var(--status-warning);
}

.channel-status.error {
  background: var(--status-error-bg);
  color: var(--status-error);
}

.channel-status.idle {
  background: var(--bg-secondary);
  color: var(--text-muted);
}

.channel-value {
  font-family: Consolas, monospace;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.progress-section {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 3px;
  padding: var(--spacing-sm) var(--spacing-md);
}

.progress-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: var(--spacing-sm);
}

.progress-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);

  .el-icon {
    font-size: 14px;
    color: var(--accent-primary);
  }
}

.progress-percent {
  color: var(--accent-primary);
  font-weight: 600;
}

.progress-bar {
  height: 4px;
  background: var(--bg-primary);
  border-radius: 2px;
  overflow: hidden;
  margin-bottom: var(--spacing-sm);
}

.progress-fill {
  height: 100%;
  background: var(--accent-primary);
  border-radius: 2px;
  transition: width 0.3s ease;
}

.progress-meta {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: var(--text-secondary);
}

@media (max-width: 1200px) {
  .channels-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}

@media (max-width: 768px) {
  .metrics-bar {
    flex-wrap: wrap;
  }

  .metric-separator {
    display: none;
  }

  .channels-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .panel-header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-sm);
  }
}
</style>
