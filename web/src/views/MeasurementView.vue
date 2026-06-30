<template>
  <PageLayout>
    <!-- ═══ 仪表盘头部 ═══ -->

    <header class="instrument-header">
      <div class="header-nav">
        <button class="back-btn" @click="goBack">
          <el-icon><ArrowLeft /></el-icon>
        </button>
      </div>

      <div class="header-identity">
        <h1 class="header-title">计量工作台</h1>
        <span class="state-chip" :class="stateClass">{{ stateLabel }}</span>
      </div>

      <div class="header-telemetry">
        <div class="telem-cell">
          <span class="telem-label">当前压力</span>
          <span class="telem-value mono">{{ displayPressure }}</span>
          <span class="telem-unit">{{ measurementStore.measureUnit || 'MPa' }}</span>
        </div>
        <span class="telem-divider" />
        <div class="telem-cell">
          <span class="telem-label">稳定性</span>
          <span class="telem-indicator" :class="measurementStore.isStable ? 'on' : 'off'">
            <span class="telem-dot" />
            {{ measurementStore.isStable ? '已稳定' : '稳定中' }}
          </span>
        </div>
        <span class="telem-divider" />
        <div class="telem-cell">
          <span class="telem-label">稳定计时</span>
          <span class="telem-value mono">{{ stableSeconds }}<small>s</small></span>
        </div>
      </div>
    </header>

    <!-- ═══ 工作台主体 ═══ -->
    <div class="workbench" :class="{ 'sidebar-collapsed': sidebarCollapsed }">
      <MeasurementSidebar
        ref="sidebarRef"
        :collapsed="sidebarCollapsed"
        @toggle="sidebarCollapsed = !sidebarCollapsed"
      />

      <main class="workbench-main">
        <div class="scroll-container">
          <MeasurementControl
            :can-start="canStart"
            :is-stable="measurementStore.isStable"
            :stable-seconds="measurementStore.stabilityState.stableDurationMs / 1000"
            :has-pressure-device="hasPressureDevice"
            @start="handleStart"
            @pause="handlePause"
            @resume="handleResume"
            @stop="handleStop"
            @retry="handleRetry"
            @restart="handleRestart"
            @reset="handleReset"
            @export="exportDialogVisible = true"
            @view-error="handleViewError"
            @manual-start="handleManualStart"
            @manual-pressurize="handleManualPressurize"
          />
          <div class="section-gap" />
          <div class="card-block">
            <MeasurementParamsPanel />
          </div>
          <div class="card-block card-block--data">
            <MeasurementDataView
              :control-mode="measurementStore.measurementParams.controlMode"
              @collect-point="handleCollectPoint"
            />
          </div>
        </div>
      </main>
    </div>

    <ExportReportDialog
      :visible="exportDialogVisible"
      :template-name="reportTemplateName"
      :point-count="measurementStore.points.length"
      :channel-count="measurementStore.channels.length"
      :pressure-mode="measurementStore.measurementParams.pressureMode"
      :exporting="isExporting"
      @close="exportDialogVisible = false"
      @export="handleExport"
    />

    <AlarmConfirmDialog
      :visible="showAlarmDialog"
      :point="alarmPoint"
      :alarm="measurementStore.alarmData"
      @decision="handleAlarmDecision"
    />
  </PageLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import { useMeasurementStore } from '@/stores/measurement'
import type { MeasurementState } from '@/stores/measurement/types'
import { useMeasurementDeviceStore } from '@/stores/measurement/deviceStore'
import { useMeasurementUI } from '@/composables/useMeasurementUI'
import { useMeasurementSync } from '@/composables/useMeasurementSync'
import { saveMeasurementAlarmConfig, exportMeasurementReport, resolveStabilityTimeout } from '@/api/measurement'
import { PressureMode } from '@/types/calibration'
import PageLayout from '@/components/common/PageLayout.vue'
import MeasurementSidebar from '@/components/measurement/MeasurementSidebar.vue'
import MeasurementControl from '@/components/measurement/MeasurementControl.vue'
import MeasurementParamsPanel from '@/components/measurement/MeasurementParamsPanel.vue'
import MeasurementDataView from '@/components/measurement/MeasurementDataView.vue'
import ExportReportDialog from '@/components/measurement/ExportReportDialog.vue'
import AlarmConfirmDialog from '@/components/measurement/AlarmConfirmDialog.vue'

const router = useRouter()
const measurementStore = useMeasurementStore()
const deviceStore = useMeasurementDeviceStore()
const measurementUI = useMeasurementUI()

const sidebarCollapsed = ref(false)
const sidebarRef = ref()
const exportDialogVisible = ref(false)
const isExporting = ref(false)

// canStart：设备连接 + 计量绑定 + 点位已生成 + 阀门=校准模式（启动必要条件）。
// valveReady / enforceValveCalibrationGate 由 measurementStore 提供，与后端门禁同源。
const canStart = computed(() =>
  deviceStore.measureDevices.some(d => d.status === 'connected') &&
  deviceStore.pressureDevices.some(d => d.status === 'connected') &&
  measurementStore.deviceBound &&
  measurementStore.points.length > 0 &&
  measurementStore.canStart
)

// startBlockedReason：返回阻断"开始采集"的具体原因文案，
// 给 UI 在按钮 tooltip / 弹框中提示用户该如何修复。
// 所有 canStart 分支都在这里穷举，调用方无需再写 `|| '兜底'` fallback。
const startBlockedReason = computed<string>(() => {
  if (!deviceStore.measureDevices.some(d => d.status === 'connected')) return '请先连接计量设备'
  if (!deviceStore.pressureDevices.some(d => d.status === 'connected')) return '请先连接压力源设备'
  if (!measurementStore.deviceBound) return '请先绑定计量设备'
  if (measurementStore.points.length === 0) return '请先生成压力表'
  if (!measurementStore.canStart) return '请先将阀门切换到校准模式'
  // 安全网：理论上 canStart=true 时该 computed 不会被读到；
  // 兜底文案让逻辑回归时仍有一条可读提示。
  return '设备状态异常，请检查后重试'
})

const hasPressureDevice = computed(() =>
  deviceStore.pressureDevices.some(d => d.status === 'connected')
)

const reportTemplateName = computed(() => {
  const count = measurementStore.points.length
  const mode = measurementStore.measurementParams.pressureMode === PressureMode.Single ? 's' : 'm'
  return `${count}点${mode === 's' ? '单程' : '回程'}模板`
})

onMounted(async () => {
  await deviceStore.loadDevices()
})

useMeasurementSync()

/* ── 状态 ── */
const STATE_LABELS: Record<MeasurementState, string> = {
  idle: '空闲',
  ready: '就绪',
  pressurizing: '打压中',
  stabilizing: '稳定中',
  collecting: '采集中',
  completed: '已完成',
  error: '错误',
  paused: '已暂停',
  stopped: '已停止'
}
const STATE_CLASSES: Record<MeasurementState, string> = {
  idle: 'chip-idle',
  ready: 'chip-running',
  pressurizing: 'chip-running',
  stabilizing: 'chip-running',
  collecting: 'chip-running',
  completed: 'chip-completed',
  error: 'chip-error',
  paused: 'chip-paused',
  stopped: 'chip-idle'
}

const stateLabel = computed(() => STATE_LABELS[measurementStore.state] || measurementStore.state)

const stateClass = computed(() => STATE_CLASSES[measurementStore.state] || '')

/* ── 遥测数据 ── */
const displayPressure = computed(() => {
  const v = measurementStore.currentPressure
  if (v === null || v === undefined) return '—'
  return Number(v).toFixed(3)
})

const stableSeconds = computed(() => {
  const ms = measurementStore.stabilityState.stableDurationMs
  return (ms / 1000).toFixed(1)
})

function goBack() { router.push('/') }

/* ── 采集控制 ── */
async function handleStart() {
  if (!canStart.value) { ElMessage.warning(startBlockedReason.value); return }
  await measurementUI.start(measurementStore.channels)
}

async function handlePause() { await measurementUI.pause() }
async function handleResume() { await measurementUI.start(measurementStore.channels) }
async function handleStop() { await measurementUI.stop() }

async function handleRetry() {
  if (!canStart.value) { ElMessage.warning(startBlockedReason.value); return }
  await measurementUI.start(measurementStore.channels)
}

async function handleRestart() {
  if (!canStart.value) { ElMessage.warning(startBlockedReason.value); return }
  measurementStore.resetCollection()
  await measurementUI.start(measurementStore.channels)
}

function handleViewError() {
  ElMessage.info('请检查设备连接和报警信息')
}

function handleReset() {
  measurementStore.resetCollection()
  ElMessage.info('采集数据已重置')
}

async function handleManualStart() {
  const hasMeasure = deviceStore.measureDevices.some(d => d.status === 'connected')
  if (!hasMeasure) { ElMessage.warning('请先连接计量设备'); return }
  if (measurementStore.points.length === 0) { ElMessage.warning('请先生成压力表'); return }
  if (!measurementStore.canStart) { ElMessage.warning('请先将阀门切换到校准模式'); return }
  await measurementUI.manualStart(measurementStore.channels)
}

async function handleManualPressurize() {
  const idx = measurementStore.currentPointIndex + 1
  await measurementUI.manualPressurize(idx)
}

async function handleCollectPoint(pointIndex: number) {
  await measurementUI.manualCollect(pointIndex)
}

/* ── 报警配置自动保存 ── */
let alarmSaveTimer: ReturnType<typeof setTimeout> | null = null
watch(
  () => [
    measurementStore.alarmConfig.enabled,
    measurementStore.alarmConfig.soundEnabled,
    measurementStore.alarmConfig.confirmOnAlarm,
    measurementStore.channels
  ],
  () => {
    if (alarmSaveTimer) clearTimeout(alarmSaveTimer)
    alarmSaveTimer = setTimeout(() => {
      const cfg = { ...measurementStore.alarmConfig }
      cfg.enabledChannels = [...measurementStore.channels]
      saveMeasurementAlarmConfig(cfg)
    }, 250)
  },
  { deep: true }
)

onUnmounted(() => {
  if (alarmSaveTimer) {
    clearTimeout(alarmSaveTimer)
    alarmSaveTimer = null
  }
})

/* ── 报警弹窗 ── */
const showAlarmDialog = computed(() =>
  measurementStore.alarmPending && measurementStore.alarmConfig.confirmOnAlarm
)

const alarmPoint = computed(() => {
  const d = measurementStore.alarmData
  if (!d) return undefined
  return measurementStore.points.find(p => p.id === d.pointId)
})

// 非确认模式：通知后自动继续采集，避免后端阻塞等待
watch(() => measurementStore.alarmPending, async (pending) => {
  if (pending && !measurementStore.alarmConfig.confirmOnAlarm && measurementStore.alarmData) {
    const a = measurementStore.alarmData
    ElMessage.warning(`报警：${a.overLimitChannels.length} 个通道精度超限，最大偏差 ${(a.maxDeviation * 100).toFixed(2)}%`)
    await measurementStore.resolveAlarm('continue')
  }
})

// 稳定超时弹窗：让用户选择继续等待或跳过当前点
watch(() => measurementStore.stabilityTimeoutPending, async (pending) => {
  if (!pending) return
  measurementStore.stabilityTimeoutPending = false
  try {
    await ElMessageBox.confirm(
      '当前压力点稳定等待超时，请选择处理方式：',
      '稳定超时',
      {
        confirmButtonText: '继续等待',
        cancelButtonText: '跳过此点',
        distinguishCancelAndClose: true,
        type: 'warning'
      }
    )
    await resolveStabilityTimeout('continue')
  } catch (action: unknown) {
    const actionStr = typeof action === 'string' ? action : ''
    if (actionStr === 'cancel') {
      await resolveStabilityTimeout('skip')
    }
  }
})

async function handleAlarmDecision(decision: 'continue' | 'recollect') {
  await measurementStore.resolveAlarm(decision)
}

/* ── 导出 ── */
async function handleExport(path: string) {
  if (!path) { ElMessage.warning('请先选择导出路径'); return }
  isExporting.value = true
  try {
    await exportMeasurementReport(path)
    ElMessage.success('报告导出成功')
    exportDialogVisible.value = false
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '报告导出失败')
  } finally {
    isExporting.value = false
  }
}
</script>

<style scoped lang="scss">
$font-sans: 'DM Sans', 'SF Pro Display', -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif;
$font-mono: 'JetBrains Mono', 'Fira Code', 'SF Mono', Consolas, monospace;
$mint: #10b981;
$mint-light: #34d399;
$mint-dark: #059669;
$slate-50: #f9fafb;
$slate-100: #f3f4f6;
$slate-200: #e5e7eb;
$slate-300: #d1d5db;
$slate-400: #9ca3af;
$slate-500: #6b7280;
$slate-600: #4b5563;
$slate-700: #374151;
$slate-800: #1f2937;
$slate-900: #111827;
$green: #22c55e;
$red: #ef4444;
$amber: #f59e0b;

/* ═══ 仪表盘头部 ═══ */
.instrument-header {
  display: flex;
  align-items: center;
  gap: 20px;
  flex-shrink: 0;
  height: 56px;
  padding: 0 24px;
  background: $slate-50;
  border-bottom: 1px solid $slate-200;
  font-family: $font-sans;
}

.header-nav { display: flex; align-items: center; }

.back-btn {
  width: 32px; height: 32px;
  background: #fff;
  border: 1px solid $slate-200;
  border-radius: 8px;
  color: $slate-500;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  transition: background 0.15s ease, color 0.15s ease, border-color 0.15s ease;

  &:hover {
    background: #fff;
    color: $mint;
    border-color: $mint;
  }
}

.header-identity {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
}

.header-title {
  font-size: 20px;
  font-weight: 600;
  color: $slate-800;
  margin: 0;
  font-family: $font-sans;
}

.header-telemetry {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-left: auto;
}

.telem-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.telem-label {
  font-size: 12px;
  font-weight: 500;
  color: $slate-400;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  white-space: nowrap;
}

.telem-value {
  font-size: 14px;
  font-weight: 600;
  color: $slate-800;
  &.mono { font-family: $font-mono; }
  small { font-size: 10px; color: $slate-400; margin-left: 1px; }
}

.telem-unit {
  font-size: 12px;
  color: $slate-400;
  font-weight: 500;
}

.telem-divider {
  width: 1px;
  height: 24px;
  background: $slate-200;
}

.telem-indicator {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 500;
  padding: 2px 8px;
  border-radius: 4px;

  &.on { color: $mint-dark; background: rgba(16, 185, 129, 0.08); }
  &.off { color: $amber; background: rgba(245, 158, 11, 0.08); }
}

.telem-dot {
  width: 6px; height: 6px; border-radius: 50%;
  .on & { background: $mint; box-shadow: 0 0 4px rgba(16, 185, 129, 0.4); animation: pulse-dot 2s ease-in-out infinite; }
  .off & { background: $amber; box-shadow: 0 0 4px rgba(245, 158, 11, 0.4); animation: pulse-dot 1.2s ease-in-out infinite; }
}

@keyframes pulse-dot {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.5; transform: scale(0.85); }
}

/* ── 状态芯片 ── */
.state-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  letter-spacing: 0.05em;

  .chip-dot { width: 5px; height: 5px; border-radius: 50%; }

  &.chip-idle { background: rgba(156,163,175,0.1); color: $slate-500; .chip-dot { background: $slate-400; } }
  &.chip-preparing, &.chip-running { background: rgba(16,185,129,0.08); color: $mint-dark; .chip-dot { background: $mint; box-shadow: 0 0 4px rgba(16,185,129,0.3); } }
  &.chip-paused { background: rgba(245,158,11,0.08); color: $amber; .chip-dot { background: $amber; } }
  &.chip-completed { background: rgba(16,185,129,0.08); color: $mint-dark; .chip-dot { background: $mint; } }
  &.chip-error { background: rgba(239,68,68,0.08); color: $red; .chip-dot { background: $red; } }
}

/* ═══ 工作台 ═══ */
.workbench {
  flex: 1;
  min-height: 0;
  display: grid;
  grid-template-columns: 280px 1fr;
  gap: 16px;
  overflow: hidden;
  position: relative;
  padding: 4px 24px 24px;
  transition: grid-template-columns 250ms cubic-bezier(0.4, 0, 0.2, 1);

  &.sidebar-collapsed {
    grid-template-columns: 32px 1fr;
  }

  &::before {
    content: '';
    position: absolute;
    inset: 0;
    background-image:
      linear-gradient(rgba(16, 185, 129, 0.04) 1px, transparent 1px),
      linear-gradient(90deg, rgba(16, 185, 129, 0.04) 1px, transparent 1px);
    background-size: 24px 24px;
    pointer-events: none;
    z-index: 0;
  }
}

.workbench-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  position: relative;
  z-index: 1;
}

.scroll-container {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 4px 4px 4px 0;

  &::-webkit-scrollbar { width: 4px; }
  &::-webkit-scrollbar-thumb { background: $slate-300; border-radius: 2px; }
  &::-webkit-scrollbar-track { background: transparent; }
}

.section-gap {
  flex-shrink: 0;
  height: 8px;
}

.card-block {
  background: #ffffff;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06), 0 1px 2px rgba(0, 0, 0, 0.04);
  position: relative;
  overflow: hidden;
  animation: card-enter 0.35s ease both;
}

.card-block--data {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

@keyframes card-enter {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}

/* 响应式 */
@media (max-width: 1024px) {
  .header-telemetry { gap: 10px; }
  .telem-label { display: none; }
}

@media (max-width: 768px) {
  .instrument-header { height: auto; flex-wrap: wrap; padding: 10px 16px; gap: 8px; }
  .header-telemetry { flex-wrap: wrap; margin-left: 0; width: 100%; }
  .telem-divider { display: none; }
  .workbench { flex-direction: column; }
}
</style>
