<template>
  <section class="realtime-data-panel">
    <header class="panel-header">
      <div class="header-title">
        <el-icon class="panel-icon">
          <DataLine />
        </el-icon>
        <h2>实时数据监控</h2>
        <span
          class="connection-status"
          :class="connectionStatusClass"
        >
          <el-icon v-if="isConnected"><CircleCheck /></el-icon>
          <el-icon v-else><CircleClose /></el-icon>
          {{ connectionStatusText }}
        </span>
      </div>
      <div class="header-actions">
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

    <div class="metrics-grid">
      <!-- 主压力值显示 -->
      <div class="metric-card primary-metric">
        <div class="metric-header">
          <span class="metric-label">当前压力</span>
          <el-icon><Odometer /></el-icon>
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
        <div class="metric-header">
          <span class="metric-label">稳定状态</span>
          <el-icon v-if="isStable">
            <CircleCheckFilled />
          </el-icon>
          <el-icon v-else>
            <WarningFilled />
          </el-icon>
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
        <div class="metric-header">
          <span class="metric-label">目标压力</span>
          <el-icon><Aim /></el-icon>
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
        <div class="metric-header">
          <span class="metric-label">环境温度</span>
          <el-icon><Thermometer /></el-icon>
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
import {
  DataLine,
  CircleCheck,
  CircleClose,
  Refresh,
  Odometer,
  CircleCheckFilled,
  WarningFilled,
  Aim,
  Grid,
  Histogram
} from '@element-plus/icons-vue'

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

<style scoped lang="scss">
/* 主面板容器 */
.realtime-data-panel {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: var(--spacing-lg);
}

/* 面板头部 */
.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-lg);
  padding-bottom: var(--spacing-md);
  border-bottom: 1px solid var(--border-color);
}

.header-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

.panel-icon {
  font-size: 24px;
  color: var(--accent-primary);
}

.header-title h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
}

.connection-status {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);
  padding: 4px 10px;
  border-radius: var(--radius-sm);
  font-size: 12px;
  font-weight: 600;
  
  .el-icon {
    font-size: 12px;
  }
}

.status-connected {
  background: rgba(16, 185, 129, 0.2);
  color: var(--status-success);
}

.status-disconnected {
  background: rgba(239, 68, 68, 0.2);
  color: var(--status-error);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

.update-time {
  font-size: 13px;
  color: var(--text-muted);
  font-family: 'SF Mono', Monaco, monospace;
}

.icon-btn {
  width: 32px;
  height: 32px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  transition: all 0.2s ease;
  
  .el-icon {
    font-size: 16px;
  }
  
  &:hover {
    background: var(--bg-secondary);
    border-color: var(--accent-primary);
    color: var(--accent-primary);
  }
}

/* 指标网格 */
.metrics-grid {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 1fr;
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-lg);
}

.metric-card {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
}

.primary-metric {
  border-color: var(--accent-primary);
}

.metric-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-sm);
  
  .el-icon {
    font-size: 16px;
    color: var(--accent-primary);
  }
}

.metric-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.metric-value-row {
  display: flex;
  align-items: baseline;
  gap: var(--spacing-xs);
  margin-bottom: var(--spacing-md);
}

.metric-value-row.small {
  margin-bottom: var(--spacing-sm);
}

.metric-value {
  font-family: 'SF Mono', 'Consolas', monospace;
  font-size: 32px;
  font-weight: 600;
  color: var(--text-primary);
  line-height: 1;
}

.metric-value-row.small .metric-value {
  font-size: 24px;
}

.metric-unit {
  font-size: 14px;
  color: var(--text-muted);
  font-weight: 500;
}

/* 压力进度条 */
.pressure-bar-container {
  margin-top: var(--spacing-md);
}

.pressure-bar-track {
  height: 8px;
  background: var(--bg-secondary);
  border-radius: 4px;
  overflow: hidden;
}

.pressure-bar-fill {
  height: 100%;
  border-radius: 4px;
  transition: width 0.3s ease, background-color 0.3s ease;
}

.bar-normal {
  background: var(--status-info);
}

.bar-approaching {
  background: var(--status-warning);
}

.bar-stable {
  background: var(--status-success);
}

.pressure-scale {
  display: flex;
  justify-content: space-between;
  margin-top: var(--spacing-xs);
  font-size: 11px;
  color: var(--text-muted);
}

/* 稳定状态显示 */
.status-display {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-md);
}

.status-indicator {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  transition: background-color 0.3s ease;
}

.indicator-stable {
  background: var(--status-success);
  box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.3);
}

.indicator-unstable {
  background: var(--status-warning);
}

.status-text {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.stability-details {
  border-top: 1px solid var(--border-color);
  padding-top: var(--spacing-md);
}

.detail-row {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  margin-bottom: var(--spacing-xs);
}

.detail-row:last-child {
  margin-bottom: 0;
}

.detail-row span {
  color: var(--text-muted);
}

.detail-row strong {
  color: var(--text-secondary);
  font-weight: 600;
}

/* 目标压力差值 */
.target-diff {
  font-size: 13px;
  font-weight: 500;
  padding: 4px 8px;
  border-radius: var(--radius-sm);
  display: inline-block;
}

.diff-ok {
  background: rgba(16, 185, 129, 0.2);
  color: var(--status-success);
}

.diff-warning {
  background: rgba(245, 158, 11, 0.2);
  color: var(--status-warning);
}

.diff-error {
  background: rgba(239, 68, 68, 0.2);
  color: var(--status-error);
}

.metric-note {
  font-size: 12px;
  color: var(--text-muted);
  font-style: italic;
}

/* 通道数据区域 */
.channels-section {
  margin-bottom: var(--spacing-lg);
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-md);
}

.section-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  
  .el-icon {
    font-size: 18px;
    color: var(--accent-primary);
  }
}

.section-header h3 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.channel-count {
  font-size: 12px;
  color: var(--text-secondary);
  background: var(--bg-tertiary);
  padding: 2px 8px;
  border-radius: var(--radius-sm);
}

.channels-grid {
  display: grid;
  grid-template-columns: repeat(8, 1fr);
  gap: var(--spacing-sm);
}

.channel-item {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  padding: var(--spacing-sm);
  text-align: center;
}

.channel-item.channel-active {
  border-color: var(--accent-primary);
  background: rgba(233, 69, 96, 0.1);
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
  color: var(--text-muted);
}

.channel-status {
  font-size: 10px;
  padding: 1px 4px;
  border-radius: 2px;
  font-weight: 500;
}

.channel-status.ok {
  background: rgba(16, 185, 129, 0.2);
  color: var(--status-success);
}

.channel-status.warning {
  background: rgba(245, 158, 11, 0.2);
  color: var(--status-warning);
}

.channel-status.error {
  background: rgba(239, 68, 68, 0.2);
  color: var(--status-error);
}

.channel-status.idle {
  background: var(--bg-secondary);
  color: var(--text-muted);
}

.channel-value {
  font-family: 'SF Mono', 'Consolas', monospace;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

/* 进度区域 */
.progress-section {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
}

.progress-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: var(--spacing-sm);
}

.progress-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  
  .el-icon {
    font-size: 16px;
    color: var(--accent-primary);
  }
}

.progress-percent {
  color: var(--accent-primary);
  font-weight: 600;
}

.progress-bar {
  height: 6px;
  background: var(--bg-secondary);
  border-radius: 3px;
  overflow: hidden;
  margin-bottom: var(--spacing-sm);
}

.progress-fill {
  height: 100%;
  background: var(--accent-primary);
  border-radius: 3px;
  transition: width 0.3s ease;
}

.progress-meta {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: var(--text-secondary);
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
    gap: var(--spacing-sm);
  }
}
</style>