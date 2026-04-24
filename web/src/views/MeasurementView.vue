<template>
  <section class="module-page">
    <div class="desktop-shell">
      <header class="module-header">
        <nav class="module-switch">
          <RouterLink class="switch-btn" :to="{ name: 'module-hub' }">
            <el-icon><ArrowLeft /></el-icon>返回
          </RouterLink>
          <RouterLink class="switch-btn" :to="{ name: 'module-device-management' }">
            设备管理
          </RouterLink>
          <RouterLink class="switch-btn active" :to="{ name: 'module-measurement' }">
            计量工作台
          </RouterLink>
          <RouterLink class="switch-btn" :to="{ name: 'module-calibration' }">
            标定模块
          </RouterLink>
          <RouterLink class="switch-btn" :to="{ name: 'module-multi-pressure' }">
            多设备打压
          </RouterLink>
        </nav>
      </header>

      <div class="workbench-layout">
        <MeasurementSidebar
          :collapsed="sidebarCollapsed"
          @toggle="sidebarCollapsed = !sidebarCollapsed"
        />

        <main class="workbench">
          <UnitConsistencyIndicator
            v-if="showUnitWarning"
            :consistency="unitConsistency"
          />

          <section class="config-bar">
            <div class="config-group">
              <label>最小值</label>
              <el-input-number v-model="uiParams.minPressure" :precision="2" :step="0.1" size="small" />
            </div>
            <div class="config-group">
              <label>最大值</label>
              <el-input-number v-model="uiParams.maxPressure" :precision="2" :step="0.1" size="small" />
            </div>
            <div class="config-group">
              <label>测点数</label>
              <el-input-number v-model="uiParams.pointCount" :min="2" :max="20" size="small" />
            </div>
            <div class="config-group">
              <label>精度</label>
              <el-input-number v-model="uiParams.precision" :min="0" :max="6" size="small" />
            </div>
            <div class="config-group">
              <label>平均次数</label>
              <el-input-number v-model="uiParams.averageCount" :min="1" :max="20" size="small" />
            </div>
            <div class="config-group">
              <label>稳定时间</label>
              <el-select v-model="uiParams.stableWaitS" size="small">
                <el-option label="1s" :value="1" />
                <el-option label="3s" :value="3" />
                <el-option label="5s" :value="5" />
                <el-option label="10s" :value="10" />
              </el-select>
            </div>
            <div class="config-group">
              <label>精度 Level</label>
              <el-select v-if="!customPrecisionMode" v-model="uiParams.precisionLevel" size="small">
                <el-option v-for="v in [0.02, 0.05, 0.1, 0.2]" :key="v" :label="`${v.toFixed(2)}%`" :value="v" />
              </el-select>
              <el-input-number v-else v-model="uiParams.precisionLevel" :min="0.0001" :max="5" :step="0.0001" :precision="4" size="small" />
              <el-checkbox v-model="customPrecisionMode" size="small">自定义</el-checkbox>
            </div>
            <div class="config-actions">
              <el-button type="primary" size="small" @click="generatePressureTable">生成压力表</el-button>
            </div>
          </section>

          <section class="control-bar">
            <div class="control-left">
              <el-radio-group v-model="uiParams.controlMode" size="small">
                <el-radio-button value="auto">自动</el-radio-button>
                <el-radio-button value="manual">手动</el-radio-button>
              </el-radio-group>
              <el-radio-group v-model="uiParams.pressureMode" size="small">
                <el-radio-button value="single">单程</el-radio-button>
                <el-radio-button value="roundTrip">回程</el-radio-button>
              </el-radio-group>
              <div class="progress-group" v-if="totalPoints > 0">
                <span class="progress-text">{{ completedPoints }}/{{ totalPoints }} 点</span>
                <el-progress :percentage="progressPercent" :stroke-width="8" size="small" />
              </div>
            </div>
            <div class="control-center">
              <span class="state-label" :class="stateClass">{{ stateLabel }}</span>
              <span class="stable-badge" :class="measurementStore.isStable ? 'is-stable' : 'is-unstable'">
                {{ measurementStore.isStable ? '稳定' : '未稳定' }}
              </span>
              <span class="stable-time">{{ stableSeconds.toFixed(1) }}s</span>
              <el-checkbox v-model="alarmCfg.enabled" size="small">报警</el-checkbox>
              <el-button size="small" @click="showChannelSelect = true">通道</el-button>
            </div>
            <div class="control-right">
              <el-button size="small" type="success" :disabled="!canStart" @click="startCollection">&#9654; 开始</el-button>
              <el-button size="small" :disabled="!isCollectingState" @click="pauseCollection">⏸ 暂停</el-button>
              <el-button size="small" :disabled="!isPausedState" @click="resumeCollection">▶ 恢复</el-button>
              <el-button size="small" type="danger" :disabled="!isActiveState" @click="stopCollection">⏹ 停止</el-button>
              <el-button v-if="measurementStore.totalRows > 0" size="small" @click="exportCSV">导出 CSV</el-button>
            </div>
          </section>

          <section class="data-section">
            <div class="data-table-wrapper">
              <table class="data-table">
                <thead>
                  <tr>
                    <th class="col-index">#</th>
                    <th class="col-status">状态</th>
                    <th class="col-target">目标</th>
                    <th v-for="ch in channelList" :key="ch" class="col-channel">CH{{ ch }}</th>
                    <th class="col-time">时间</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="point in pointRows" :key="point.id"
                    :class="getRowClass(point)"
                  >
                    <td>{{ point.index }}</td>
                    <td>
                      <span class="status-tag" :class="point.status">{{ statusLabel(point.status) }}</span>
                    </td>
                    <td class="cell-num">{{ point.targetPressure.toFixed(precision) }}</td>
                    <td v-for="ch in channelList" :key="ch" class="cell-num"
                      :class="getChannelValueClass(point, ch)"
                    >
                      {{ getChannelValue(point, ch) }}
                    </td>
                    <td class="cell-time">{{ point.collectTime ? formatTime(point.collectTime) : '--' }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>
        </main>
      </div>

      <AlarmConfirmDialog
        v-model:visible="alarmDialogVisible"
        :alarm="measurementStore.alarmData"
        @decision="handleAlarmDecision"
      />
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { ArrowLeft } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

import AlarmConfirmDialog from '@/components/measurement/AlarmConfirmDialog.vue'
import UnitConsistencyIndicator from '@/components/common/UnitConsistencyIndicator.vue'
import MeasurementSidebar from '@/components/measurement/MeasurementSidebar.vue'
import { useMeasurementStore } from '@/stores/measurement'
import { useMeasurementDeviceStore } from '@/stores/measurement/deviceStore'
import { fetchUnitConsistency } from '@/api/device'

const measurementStore = useMeasurementStore()
const deviceStore = useMeasurementDeviceStore()

// ── 布局状态 ──
const sidebarCollapsed = ref(false)

// ── 参数配置 ──
const customPrecisionMode = ref(false)
const precision = computed(() => measurementStore.config?.precision ?? 3)
const selectedChannels = ref<number[]>([])
const channelList = computed(() => selectedChannels.value.length > 0 ? selectedChannels.value : [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16])

interface UiParams {
  minPressure: number; maxPressure: number; pointCount: number; precision: number
  averageCount: number; stableWaitS: number; precisionLevel: number
  controlMode: 'auto' | 'manual'; pressureMode: 'single' | 'roundTrip'
}

const uiParams = ref<UiParams>({
  minPressure: 0, maxPressure: 100, pointCount: 5, precision: 3,
  averageCount: 3, stableWaitS: 5, precisionLevel: 0.02,
  controlMode: 'auto', pressureMode: 'single'
})

const alarmCfg = ref({ enabled: true, confirm: true, sound: false })

// ── 单位一致性 ──
const unitConsistency = ref<{ consistent: boolean; message: string; units: { deviceName: string; unit: string }[] }>({
  consistent: true, message: '', units: []
})
const showUnitWarning = computed(() => !unitConsistency.value.consistent)

// ── 状态与进度 ──
const stateLabel = computed(() => {
  const labels: Record<string, string> = {
    idle: '空闲', ready: '就绪', pressuring: '打压中', stabilizing: '稳定中',
    collecting: '采集中', completed: '已完成', error: '错误', paused: '已暂停'
  }
  return labels[measurementStore.state] || measurementStore.state
})

const stateClass = computed(() => {
  const m = measurementStore.state
  if (m === 'collecting') return 'is-collecting'
  if (m === 'paused') return 'is-paused'
  if (m === 'error') return 'is-error'
  if (m === 'completed') return 'is-completed'
  return ''
})

const isCollectingState = computed(() => measurementStore.isRunning || measurementStore.isCollecting)
const isPausedState = computed(() => measurementStore.isPaused)
const isActiveState = computed(() => !measurementStore.isIdle)
const stableSeconds = computed(() => Number((measurementStore.stabilityState.stableDurationMs / 1000).toFixed(1)))

const canStart = computed(() =>
  measurementStore.isStartable &&
  measurementStore.deviceBound &&
  selectedChannels.value.length > 0 &&
  unitConsistency.value.consistent
)

// ── 测点表 ──
interface PointRow {
  id: string; index: number; targetPressure: number; status: string
  collectedData?: number[]; collectTime?: string; actualPressure?: number
}

const precisionDisplay = computed(() => {
  const p = uiParams.value.precision
  return p > 0 ? p : 2
})

const pointRows = computed<PointRow[]>(() =>
  measurementStore.points.map(p => ({
    id: p.id,
    index: p.index,
    targetPressure: p.targetPressure,
    status: p.status,
    collectedData: p.collectedData,
    collectTime: p.collectTime,
    actualPressure: p.actualPressure
  }))
)

const totalPoints = computed(() => measurementStore.points.length)
const completedPoints = computed(() => measurementStore.points.filter(p => p.status === 'completed').length)
const progressPercent = computed(() =>
  totalPoints.value > 0 ? Math.round((completedPoints.value / totalPoints.value) * 100) : 0
)

// ── 报警 ──
const alarmDialogVisible = ref(false)
const showChannelSelect = ref(false)

watch(() => measurementStore.alarmPending, (val) => { if (val) alarmDialogVisible.value = true })

function getRowClass(point: PointRow) {
  return {
    'row-current': point.index === completedPoints.value + 1,
    'row-completed': point.status === 'completed',
    'row-error': point.status === 'error'
  }
}

function statusLabel(s: string) {
  const map: Record<string, string> = {
    pending: '待执行', pressurizing: '打压中', stabilizing: '稳定中',
    collecting: '采集中', completed: '已完成', error: '异常'
  }
  return map[s] || s
}

function getChannelValue(point: PointRow, ch: number): string {
  if (!point.collectedData || ch > point.collectedData.length) return '--'
  return point.collectedData[ch - 1]?.toFixed(precisionDisplay.value) ?? '--'
}

function getChannelValueClass(point: PointRow, ch: number) {
  if (!point.collectedData || ch > point.collectedData.length) return ''
  const val = point.collectedData[ch - 1]
  const target = point.targetPressure
  if (target === 0) return ''
  const dev = Math.abs(val - target) / target
  if (dev > (measurementStore.alarmConfig?.threshold ?? 0.02)) return 'cell-warn'
  return ''
}

function formatTime(t: string) {
  if (!t) return '--'
  return new Date(t).toLocaleTimeString()
}

// ── 操作方法 ──
async function checkUnitConsistency() {
  try {
    const result = await fetchUnitConsistency()
    unitConsistency.value = {
      consistent: result.consistent,
      message: result.conflicts?.join('；') || '',
      units: result.conflicts?.map((c: string) => ({ deviceName: c, unit: '' })) || []
    }
  } catch { /* silent */ }
}

async function generatePressureTable() {
  const cfg = measurementStore.config
  if (cfg) {
    await measurementStore.saveConfig({
      minPressure: uiParams.value.minPressure,
      maxPressure: uiParams.value.maxPressure,
      pointCount: uiParams.value.pointCount,
      precision: uiParams.value.precision,
      averageCount: uiParams.value.averageCount,
      stableDurationMs: uiParams.value.stableWaitS * 1000,
      precisionLevel: uiParams.value.precisionLevel,
      pressureMode: uiParams.value.pressureMode,
      controlMode: uiParams.value.controlMode
    })
  }
  await measurementStore.generatePoints()
}

async function startCollection() {
  if (!measurementStore.deviceBound) {
    ElMessage.warning('请先连接设备')
    return
  }
  if (selectedChannels.value.length === 0) {
    ElMessage.warning('请选择采集通道')
    return
  }
  await measurementStore.start(selectedChannels.value)
}

function pauseCollection() { measurementStore.pause() }
function resumeCollection() { measurementStore.start(selectedChannels.value) }
function stopCollection() { measurementStore.stop() }

async function handleAlarmDecision(decision: 'continue' | 'retry') {
  await measurementStore.resolveAlarm(decision)
  alarmDialogVisible.value = false
}

// ── CSV 导出 ──
function exportCSV() {
  const rows = measurementStore.rows
  if (rows.length === 0) { ElMessage.warning('没有可导出的数据'); return }
  const headers = ['时间', ...channelList.value.map(ch => `CH${ch}`)]
  const csvRows = rows.map(r => [
    r.timestamp,
    ...channelList.value.map(ch => r.channels[String(ch)]?.toFixed(4) ?? '')
  ])
  const csvContent = [headers.join(','), ...csvRows.map(r => r.join(','))].join('\n')
  const blob = new Blob(['\ufeff' + csvContent], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  link.href = URL.createObjectURL(blob)
  link.download = `measurement_${new Date().toISOString().split('T')[0]}.csv`
  link.click()
  ElMessage.success('CSV 已导出')
}

// ── 生命周期 ──
onMounted(async () => {
  await Promise.all([deviceStore.loadDevices(), measurementStore.fetchCurrentState()])
  if (measurementStore.channels.length > 0) selectedChannels.value = [...measurementStore.channels]
  const connectedMeasure = deviceStore.measureDevices.find(d => d.status === 'connected')
  const connectedPressure = deviceStore.pressureDevices.find(d => d.status === 'connected')
  if (connectedMeasure && connectedPressure) {
    await measurementStore.bindDevices(connectedMeasure.id, connectedPressure.id)
  } else if (connectedMeasure) {
    await measurementStore.bindMeasureDevice(connectedMeasure.id)
  }
  if (connectedMeasure) {
    await Promise.all([
      measurementStore.refreshDeviceInfo(),
      measurementStore.refreshValveStatus(),
      measurementStore.refreshMeasureUnit()
    ])
  }
  await checkUnitConsistency()
  measurementStore.setupSSE()
  startPolling()
  startDeviceRefresh()
})

onUnmounted(() => {
  measurementStore.teardownSSE()
  stopPolling()
})

let pollTimer: ReturnType<typeof setInterval> | null = null
let deviceRefreshTimer: ReturnType<typeof setInterval> | null = null

function startPolling() {
  if (pollTimer) return
  pollTimer = setInterval(async () => {
    if (measurementStore.isRunning) {
      await Promise.all([measurementStore.refreshPressure(), measurementStore.refreshStability()])
    }
  }, 2000)
}

function startDeviceRefresh() {
  if (deviceRefreshTimer) return
  deviceRefreshTimer = setInterval(() => { deviceStore.loadDevices(true) }, 5000)
}

function stopPolling() {
  if (pollTimer) { clearInterval(pollTimer); pollTimer = null }
  if (deviceRefreshTimer) { clearInterval(deviceRefreshTimer); deviceRefreshTimer = null }
}
</script>

<style scoped lang="scss">
.module-page {
  padding: var(--spacing-lg);
  box-sizing: border-box;
  height: 100%;
  overflow: hidden;
  background: var(--bg-primary);
}
.desktop-shell {
  max-width: 100%;
  height: 100%;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}
.module-header {
  align-items: flex-end;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  gap: var(--spacing-lg);
  justify-content: space-between;
  padding-bottom: var(--spacing-lg);
  flex-shrink: 0;
  min-height: 52px;
}
.module-switch {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-xs);
}
.switch-btn {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  padding: 6px 14px;
  text-decoration: none;
  font-size: 13px;
  transition: all 0.15s ease;
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);
  &:hover { background: var(--bg-quaternary); color: var(--text-primary); }
  .el-icon { font-size: 12px; }
  &.active { background: var(--accent-primary); border-color: var(--accent-primary); color: var(--bg-primary); font-weight: 600; }
}

// 布局
.workbench-layout {
  flex: 1;
  min-height: 0;
  display: flex;
  gap: 0;
  background: var(--bg-primary);
}

// 侧边栏
.sidebar {
  width: 320px;
  background: var(--bg-secondary);
  border-right: 1px solid var(--border-color);
  position: relative;
  transition: width 0.25s ease;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  &.collapsed { width: 32px; }
}
.sidebar-toggle {
  position: absolute;
  right: -12px;
  top: 50%;
  transform: translateY(-50%);
  width: 12px; height: 36px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-left: none;
  border-radius: 0 var(--radius-sm) var(--radius-sm) 0;
  display: flex; align-items: center; justify-content: center;
  cursor: pointer; z-index: 10;
  .el-icon { color: var(--text-secondary); font-size: 10px; }
  &:hover { background: var(--bg-quaternary); .el-icon { color: var(--accent-primary); } }
}
.sidebar-body {
  padding: var(--spacing-md);
  overflow-y: auto;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}
.sidebar-section {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}
.section-title {
  color: var(--text-secondary);
  font-size: 11px;
  font-weight: 600;
  margin: 0;
  text-transform: uppercase;
  letter-spacing: 0.06em;
}
.empty-device-hint {
  padding: var(--spacing-md) 0;
  text-align: center;
}

// 工作区
.workbench {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
  padding: var(--spacing-md) var(--spacing-lg);
  overflow-y: auto;
  overflow-x: hidden;
}

// 配置条
.config-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  gap: var(--spacing-sm) var(--spacing-md);
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-sm) var(--spacing-md);
  flex-shrink: 0;
}
.config-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
  label { color: var(--text-muted); font-size: 11px; font-weight: 500; white-space: nowrap; }
  :deep(.el-input-number) { width: 96px; }
  :deep(.el-select) { width: 96px; }
}
.config-actions {
  display: flex;
  align-items: flex-end;
  margin-left: auto;
}

// 控制条
.control-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--spacing-sm) var(--spacing-md);
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-sm) var(--spacing-md);
  flex-shrink: 0;
}
.control-left, .control-center, .control-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
}
.control-left { flex: 1; }
.control-center { gap: var(--spacing-xs); }
.control-right { gap: var(--spacing-xs); }
.progress-group {
  display: flex;
  flex-direction: column;
  min-width: 140px;
}
.progress-text {
  color: var(--text-secondary);
  font-size: 10px;
}
.state-label {
  font-size: 12px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 999px;
  background: var(--bg-secondary);
  color: var(--text-secondary);
  &.is-collecting { color: var(--status-info); }
  &.is-paused { color: var(--status-warning); }
  &.is-error { color: var(--status-error); }
  &.is-completed { color: var(--status-success); }
}
.stable-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 999px;
  font-weight: 600;
  &.is-stable { background: var(--status-success-bg); color: var(--status-success); }
  &.is-unstable { background: var(--status-warning-bg); color: var(--status-warning); }
}
.stable-time {
  color: var(--text-secondary);
  font-size: 11px;
  font-variant-numeric: tabular-nums;
}

// 数据表格
.data-section {
  flex: 1;
  min-height: 0;
  overflow: auto;
}
.data-table-wrapper { height: 100%; overflow: auto; }
.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
  th {
    background: var(--bg-tertiary);
    color: var(--text-muted);
    font-weight: 600;
    padding: 6px 8px;
    text-align: right;
    white-space: nowrap;
    position: sticky;
    top: 0;
    z-index: 1;
    border-bottom: 1px solid var(--border-color);
    &.col-index, &.col-status { text-align: center; }
    &.col-target { text-align: right; }
    &.col-time { text-align: center; }
  }
  td {
    padding: 6px 8px;
    border-bottom: 1px solid var(--border-color-weak);
    color: var(--text-primary);
    text-align: right;
    white-space: nowrap;
    &.cell-num { font-variant-numeric: tabular-nums; }
    &.cell-warn { color: var(--status-warning); font-weight: 600; }
  }
  .col-index, .col-status, .col-time { text-align: center; }
}
.status-tag {
  display: inline-block;
  padding: 1px 6px;
  border-radius: 4px;
  font-size: 10px;
  font-weight: 600;
  &.pending { background: var(--bg-quaternary); color: var(--text-secondary); }
  &.pressurizing, &.stabilizing { background: var(--status-info-bg); color: var(--status-info); }
  &.collecting { background: var(--status-info-bg); color: var(--status-info); }
  &.completed { background: var(--status-success-bg); color: var(--status-success); }
  &.error { background: var(--status-error-bg); color: var(--status-error); }
}
.row-current td { background: rgba(var(--accent-primary-rgb, 6, 182, 212), 0.08); }
.row-completed td { color: var(--text-secondary); }
.row-error td { background: rgba(239, 68, 68, 0.06); }

@media (max-width: 900px) {
  .workbench-layout { flex-direction: column; }
  .sidebar { width: 100% !important; border-right: none; border-bottom: 1px solid var(--border-color); .sidebar-toggle { display: none; } }
  .config-bar { flex-direction: column; align-items: stretch; }
  .control-bar { flex-direction: column; align-items: stretch; }
  .control-left, .control-center, .control-right { flex-wrap: wrap; }
}
</style>
