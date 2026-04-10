<template>
  <section class="realtime-data-panel">
    <header class="panel-header">
      <div class="header-title">
        <h2>实时数据监控</h2>
        <span
          class="connection-status"
          :class="connectionStatusClass"
        >
          {{ connectionStatusText }}
        </span>
      </div>
      <div class="header-actions">
        <span class="update-time">更新: {{ lastUpdateTime }}</span>
        <button
          type="button"
          class="icon-btn"
          title="重新连接"
          @click="reconnect"
        >
          ↻
        </button>
      </div>
    </header>

    <div class="metrics-grid">
      <!-- 主压力值显示 -->
      <div class="metric-card primary-metric">
        <div class="metric-label">
          当前压力
        </div>
        <div class="metric-value-row">
          <span class="metric-value">
            {{ formatPressure(currentPressure) }}
          </span>
          <span class="metric-unit">kPa</span>
        </div>
        <!-- 压力可视化条 -->
        <div class="pressure-bar-container">
          <div class="pressure-bar-track">
            <div
              class="pressure-bar-fill"
              :style="{ width: pressurePercentage + '%' }"
              :class="pressureBarClass"
            />
          </div>
          <div class="pressure-scale">
            <span>0</span>
            <span>{{ targetPressure / 2 }}kPa</span>
            <span>{{ targetPressure }}kPa</span>
          </div>
        </div>
      </div>

      <!-- 稳定状态 -->
      <div class="metric-card status-metric">
        <div class="metric-label">
          稳定状态
        </div>
        <div class="status-display">
          <span
            class="status-indicator"
            :class="stabilityClass"
          />
          <span class="status-text">
            {{ stabilityText }}
          </span>
        </div>
        <div class="stability-details">
          <div class="detail-row">
            <span>持续时间</span>
            <strong>{{ stableDuration }}s</strong>
          </div>
          <div class="detail-row">
            <span>阈值</span>
            <strong>±{{ stabilityThreshold }}kPa</strong>
          </div>
        </div>
      </div>

      <!-- 目标压力 -->
      <div class="metric-card secondary-metric">
        <div class="metric-label">
          目标压力
        </div>
        <div class="metric-value-row small">
          <span class="metric-value">
            {{ formatPressure(targetPressure) }}
          </span>
          <span class="metric-unit">kPa</span>
        </div>
        <div
          class="target-diff"
          :class="targetDiffClass"
        >
          {{ targetDiffText }}
        </div>
      </div>

      <!-- 温度值 -->
      <div class="metric-card secondary-metric">
        <div class="metric-label">
          环境温度
        </div>
        <div class="metric-value-row small">
          <span class="metric-value">
            {{ formatTemperature(temperature) }}
          </span>
          <span class="metric-unit">°C</span>
        </div>
        <div class="metric-note">
          参考值
        </div>
      </div>
    </div>

    <!-- 通道数据表格 -->
    <div class="channels-section">
      <div class="section-header">
        <h3>通道读数</h3>
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
        <span>采集进度</span>
        <span>{{ progressPercent }}%</span>
      </div>
      <div class="progress-bar">
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

interface ChannelInfo {
  value: number | null
  status: 'ok' | 'warning' | 'error' | 'idle'
  isActive: boolean
}

// 响应式数据
const currentPressure = ref(0)
const targetPressure = ref(1000)
const temperature = ref(25.5)
const isStable = ref(false)
const stableDuration = ref(0)
const stabilityThreshold = ref(0.5)
const lastUpdateTime = ref('--:--:--')
const isConnected = ref(false)

const channelData = ref<ChannelInfo[]>(
  Array.from({ length: 16 }, () => ({
    value: null,
    status: 'idle',
    isActive: false
  }))
)

const showProgress = ref(true)
const completedPoints = ref(3)
const totalPoints = ref(10)
const estimatedTime = ref('12分钟')

// 计算属性
const connectionStatusClass = computed(() => ({
  'status-connected': isConnected.value,
  'status-disconnected': !isConnected.value
}))

const connectionStatusText = computed(() =>
  isConnected.value ? '已连接' : '未连接'
)

const pressurePercentage = computed(() => {
  if (!targetPressure.value) return 0
  const percent = (currentPressure.value / targetPressure.value) * 100
  return Math.min(Math.max(percent, 0), 100)
})

const pressureBarClass = computed(() => {
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

const targetDiff = computed(() =>
  currentPressure.value - targetPressure.value
)

const targetDiffClass = computed(() => {
  const diff = Math.abs(targetDiff.value)
  if (diff <= stabilityThreshold.value) return 'diff-ok'
  if (diff <= stabilityThreshold.value * 3) return 'diff-warning'
  return 'diff-error'
})

const targetDiffText = computed(() => {
  const diff = targetDiff.value
  const sign = diff > 0 ? '+' : ''
  return `${sign}${formatPressure(diff)} kPa`
})

const activeChannels = computed(() =>
  channelData.value.filter(ch => ch.isActive).length
)

const totalChannels = computed(() => 16)

const progressPercent = computed(() =>
  Math.round((completedPoints.value / totalPoints.value) * 100)
)

// 方法
function formatPressure(value: number | null): string {
  if (value === null) return '--.--'
  return value.toFixed(2)
}

function formatTemperature(value: number | null): string {
  if (value === null) return '--.-'
  return value.toFixed(1)
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

function reconnect() {
  // TODO: 实现重连逻辑
  console.log('重新连接设备...')
}

// 模拟实时数据更新
let updateInterval: ReturnType<typeof setInterval> | null = null

function startSimulation() {
  updateInterval = setInterval(() => {
    // 更新时间
    const now = new Date()
    lastUpdateTime.value = now.toLocaleTimeString('zh-CN')

    // 模拟压力值波动
    const target = targetPressure.value
    const variation = (Math.random() - 0.5) * 2
    currentPressure.value = Math.max(0, target + variation)

    // 模拟稳定状态
    isStable.value = Math.abs(variation) < stabilityThreshold.value
    if (isStable.value) {
      stableDuration.value += 1
    } else {
      stableDuration.value = 0
    }

    // 模拟连接状态
    isConnected.value = true

    // 模拟通道数据
    channelData.value = channelData.value.map((ch, idx) => {
      if (idx < 8) { // 前8个通道激活
        return {
          value: currentPressure.value + (Math.random() - 0.5) * 0.1,
          status: Math.random() > 0.9 ? 'warning' : 'ok',
          isActive: true
        }
      }
      return ch
    })
  }, 1000)
}

function stopSimulation() {
  if (updateInterval) {
    clearInterval(updateInterval)
    updateInterval = null
  }
}

// 生命周期
onMounted(() => {
  startSimulation()
})

onUnmounted(() => {
  stopSimulation()
})
</script>

<style scoped>
/* 主面板容器 */
.realtime-data-panel {
  background: #ffffff;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  padding: 20px;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

/* 面板头部 */
.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid #e5e7eb;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-title h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #111827;
}

.connection-status {
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.status-connected {
  background: #d1fae5;
  color: #065f46;
}

.status-disconnected {
  background: #fee2e2;
  color: #991b1b;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.update-time {
  font-size: 13px;
  color: #6b7280;
  font-family: 'SF Mono', Monaco, monospace;
}

.icon-btn {
  width: 32px;
  height: 32px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  background: #ffffff;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  transition: all 0.15s ease;
}

.icon-btn:hover {
  background: #f3f4f6;
  border-color: #9ca3af;
}

/* 指标网格 */
.metrics-grid {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 1fr;
  gap: 16px;
  margin-bottom: 24px;
}

.metric-card {
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  padding: 16px;
}

.primary-metric {
  background: #f0f9ff;
  border-color: #bae6fd;
}

.metric-label {
  font-size: 12px;
  font-weight: 500;
  color: #6b7280;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 8px;
}

.metric-value-row {
  display: flex;
  align-items: baseline;
  gap: 6px;
  margin-bottom: 12px;
}

.metric-value-row.small {
  margin-bottom: 8px;
}

.metric-value {
  font-family: 'SF Mono', 'Consolas', monospace;
  font-size: 32px;
  font-weight: 600;
  color: #111827;
  line-height: 1;
}

.metric-value-row.small .metric-value {
  font-size: 24px;
}

.metric-unit {
  font-size: 14px;
  color: #6b7280;
  font-weight: 500;
}

/* 压力进度条 */
.pressure-bar-container {
  margin-top: 12px;
}

.pressure-bar-track {
  height: 8px;
  background: #e5e7eb;
  border-radius: 4px;
  overflow: hidden;
}

.pressure-bar-fill {
  height: 100%;
  border-radius: 4px;
  transition: width 0.3s ease, background-color 0.3s ease;
}

.bar-normal {
  background: #3b82f6;
}

.bar-approaching {
  background: #f59e0b;
}

.bar-stable {
  background: #10b981;
}

.pressure-scale {
  display: flex;
  justify-content: space-between;
  margin-top: 4px;
  font-size: 11px;
  color: #9ca3af;
}

/* 稳定状态显示 */
.status-display {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.status-indicator {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  transition: background-color 0.3s ease;
}

.indicator-stable {
  background: #10b981;
  box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.2);
}

.indicator-unstable {
  background: #f59e0b;
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.status-text {
  font-size: 16px;
  font-weight: 600;
  color: #111827;
}

.stability-details {
  border-top: 1px solid #e5e7eb;
  padding-top: 12px;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  margin-bottom: 6px;
}

.detail-row:last-child {
  margin-bottom: 0;
}

.detail-row span {
  color: #6b7280;
}

.detail-row strong {
  color: #374151;
  font-weight: 600;
}

/* 目标压力差值 */
.target-diff {
  font-size: 13px;
  font-weight: 500;
  padding: 4px 8px;
  border-radius: 4px;
  display: inline-block;
}

.diff-ok {
  background: #d1fae5;
  color: #065f46;
}

.diff-warning {
  background: #fef3c7;
  color: #92400e;
}

.diff-error {
  background: #fee2e2;
  color: #991b1b;
}

.metric-note {
  font-size: 12px;
  color: #9ca3af;
  font-style: italic;
}

/* 通道数据区域 */
.channels-section {
  margin-bottom: 24px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.section-header h3 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: #374151;
}

.channel-count {
  font-size: 12px;
  color: #6b7280;
  background: #f3f4f6;
  padding: 2px 8px;
  border-radius: 4px;
}

.channels-grid {
  display: grid;
  grid-template-columns: repeat(8, 1fr);
  gap: 8px;
}

.channel-item {
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 4px;
  padding: 8px;
  text-align: center;
}

.channel-item.channel-active {
  background: #f0f9ff;
  border-color: #bae6fd;
}

.channel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.channel-name {
  font-size: 11px;
  font-weight: 600;
  color: #6b7280;
}

.channel-status {
  font-size: 10px;
  padding: 1px 4px;
  border-radius: 2px;
  font-weight: 500;
}

.channel-status.ok {
  background: #d1fae5;
  color: #065f46;
}

.channel-status.warning {
  background: #fef3c7;
  color: #92400e;
}

.channel-status.error {
  background: #fee2e2;
  color: #991b1b;
}

.channel-status.idle {
  background: #f3f4f6;
  color: #9ca3af;
}

.channel-value {
  font-family: 'SF Mono', 'Consolas', monospace;
  font-size: 14px;
  font-weight: 600;
  color: #111827;
}

/* 进度区域 */
.progress-section {
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  padding: 16px;
}

.progress-header {
  display: flex;
  justify-content: space-between;
  font-size: 14px;
  font-weight: 500;
  color: #374151;
  margin-bottom: 8px;
}

.progress-bar {
  height: 6px;
  background: #e5e7eb;
  border-radius: 3px;
  overflow: hidden;
  margin-bottom: 8px;
}

.progress-fill {
  height: 100%;
  background: #3b82f6;
  border-radius: 3px;
  transition: width 0.3s ease;
}

.progress-meta {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #6b7280;
}

/* 响应式适配 */
@media (max-width: 1200px) {
  .metrics-grid {
    grid-template-columns: 1fr 1fr;
  }
  
  .channels-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}

@media (max-width: 768px) {
  .metrics-grid {
    grid-template-columns: 1fr;
  }
  
  .channels-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .panel-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
}
</style>
