<template>
  <PageLayout>
    <header class="page-header">
      <div class="header-left">
        <button class="back-btn" @click="goBack">
          <el-icon><ArrowLeft /></el-icon>
        </button>
        <div class="header-title">
          <h1>计量工作台</h1>
          <span class="state-badge" :class="stateClass">{{ stateLabel }}</span>
        </div>
      </div>
      <div class="header-right">
        <div class="status-info">
          <span class="info-label">稳定:</span>
          <span
            class="info-value"
            :class="measurementStore.isStable ? 'stable' : 'unstable'"
          >
            {{ measurementStore.isStable ? '是' : '否' }}
          </span>
        </div>
        <div class="status-info">
          <span class="info-label">实时时间:</span>
          <span class="time-badge">
            {{ (measurementStore.stabilityState.stableDurationMs / 1000).toFixed(1) }}<small>s</small>
          </span>
        </div>
      </div>
    </header>

    <div class="workbench-content">
      <MeasurementSidebar
        ref="sidebarRef"
        :collapsed="sidebarCollapsed"
        @toggle="sidebarCollapsed = !sidebarCollapsed"
      />

      <main class="workbench-main">
        <MeasurementControl
          :can-start="canStart"
          :is-stable="measurementStore.isStable"
          :stable-seconds="measurementStore.stabilityState.stableDurationMs / 1000"
          @start="handleStart"
          @pause="handlePause"
          @resume="handleResume"
          @stop="handleStop"
          @reset="handleReset"
          @export="exportDialogVisible = true"
          @select-channel="channelDialogVisible = true"
          @manual-pressurize="handleManualPressurize"
          @manual-collect="handleManualCollect"
        />
        <MeasurementParamsPanel />
        <MeasurementDataView />
      </main>
    </div>

    <AlarmChannelSelectDialog
      :visible="channelDialogVisible"
      @close="channelDialogVisible = false"
      @confirm="onChannelConfirm"
    />

    <ExportReportDialog
      :visible="exportDialogVisible"
      :template-name="reportTemplateName"
      :point-count="measurementStore.points.length"
      :pressure-mode="measurementStore.measurementParams.pressureMode"
      :exporting="isExporting"
      @close="exportDialogVisible = false"
      @export="handleExport"
      @select-path="handleSelectPath"
    />
  </PageLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import { useMeasurementStore } from '@/stores/measurement'
import { useMeasurementDeviceStore } from '@/stores/measurement/deviceStore'
import { saveMeasurementAlarmConfig } from '@/api/measurement'
import PageLayout from '@/components/common/PageLayout.vue'
import MeasurementSidebar from '@/components/measurement/MeasurementSidebar.vue'
import MeasurementControl from '@/components/measurement/MeasurementControl.vue'
import MeasurementParamsPanel from '@/components/measurement/MeasurementParamsPanel.vue'
import MeasurementDataView from '@/components/measurement/MeasurementDataView.vue'
import AlarmChannelSelectDialog from '@/components/measurement/AlarmChannelSelectDialog.vue'
import ExportReportDialog from '@/components/measurement/ExportReportDialog.vue'

const router = useRouter()
const measurementStore = useMeasurementStore()
const deviceStore = useMeasurementDeviceStore()

const sidebarCollapsed = ref(false)
const sidebarRef = ref()
const channelDialogVisible = ref(false)
const exportDialogVisible = ref(false)
const isExporting = ref(false)

const canStart = computed(() =>
  deviceStore.measureDevices.some(d => d.status === 'connected') &&
  deviceStore.pressureDevices.some(d => d.status === 'connected') &&
  measurementStore.points.length > 0
)

const reportTemplateName = computed(() => {
  const count = measurementStore.points.length
  const mode = measurementStore.measurementParams.pressureMode
  return `${count}${mode === 'single' ? 's' : 'm'}.xlsx`
})

onMounted(async () => {
  await Promise.all([
    deviceStore.loadDevices(),
    measurementStore.loadAlarmConfig()
  ])
  measurementStore.setupSSE()
})

onUnmounted(() => {
  measurementStore.teardownSSE()
})

const stateLabel = computed(() => {
  const m: Record<string, string> = {
    idle: '空闲', pressuring: '打压中', stabilizing: '稳定中',
    collecting: '采集中', completed: '已完成', error: '错误', paused: '已暂停'
  }
  return m[measurementStore.state] || measurementStore.state
})

const stateClass = computed(() => {
  const m: Record<string, string> = {
    idle: 'state-idle', preparing: 'state-preparing', measuring: 'state-measuring',
    paused: 'state-paused', completed: 'state-completed', error: 'state-error'
  }
  return m[measurementStore.state] || ''
})

function goBack() { router.push('/') }

// ── 采集控制 ──

async function handleStart() {
  if (!canStart.value) { ElMessage.warning('请先连接设备并生成压力表'); return }
  await measurementStore.start(measurementStore.channels)
}

async function handlePause() { await measurementStore.pause() }
async function handleResume() { await measurementStore.start(measurementStore.channels) }
async function handleStop() { await measurementStore.stop() }

function handleReset() {
  measurementStore.resetCollection()
  ElMessage.info('采集数据已重置')
}

async function handleManualPressurize() {
  const idx = measurementStore.currentPointIndex || 1
  await measurementStore.manualPressurize(idx)
}

async function handleManualCollect() {
  const idx = measurementStore.currentPointIndex || 1
  await measurementStore.manualCollect(idx)
  measurementStore.completePoint()
}

// ── 报警通道选择 ──

function onChannelConfirm(channels: number[]) {
  measurementStore.alarmConfig.enabledChannels = channels
  channelDialogVisible.value = false
}

// ── 报警配置自动保存（250ms 防抖） ──

let alarmSaveTimer: ReturnType<typeof setTimeout> | null = null
watch(
  () => [
    measurementStore.alarmConfig.enabled,
    measurementStore.alarmConfig.soundEnabled,
    measurementStore.alarmConfig.confirmOnAlarm,
    measurementStore.alarmConfig.enabledChannels
  ],
  () => {
    if (alarmSaveTimer) clearTimeout(alarmSaveTimer)
    alarmSaveTimer = setTimeout(() => {
      saveMeasurementAlarmConfig(measurementStore.alarmConfig)
    }, 250)
  },
  { deep: true }
)

// ── 导出报告 ──

async function handleSelectPath() {
  ElMessage.info('请选择导出路径（当前尚未对接文件对话框 API）')
}

async function handleExport(path: string) {
  if (!path) { ElMessage.warning('请先选择导出路径'); return }
  isExporting.value = true
  try {
    const url = measurementStore.exportUrl
    const a = document.createElement('a')
    a.href = url
    a.download = `measurement-export-${Date.now()}.csv`
    a.click()
    ElMessage.success('报告导出成功')
    exportDialogVisible.value = false
  } catch (error) {
    ElMessage.error('报告导出失败')
  } finally {
    isExporting.value = false
  }
}
</script>

<style scoped lang="scss">
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
  height: 48px;
  padding: 0 24px;
  background: #ffffff;
  border-bottom: 1px solid #e5e7eb;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.back-btn {
  width: 28px;
  height: 28px;
  background: transparent;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  color: #9ca3af;
  cursor: pointer;
  transition: all 0.15s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;

  &:hover {
    background: #f9fafb;
    color: #4b5563;
    border-color: #d1d5db;
  }
}

.header-title {
  display: flex;
  align-items: center;
  gap: 10px;

  h1 {
    font-size: 18px;
    font-weight: 700;
    color: #1f2937;
    margin: 0;
    letter-spacing: -0.01em;
    font-family: 'Inter', 'SF Pro Display', -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif;
  }
}

/* 状态徽章：按 Tags 规范 — 15% 背景透明度，30% 边框透明度 */
.state-badge {
  padding: 3px 10px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  line-height: 1.5;
  letter-spacing: 0.02em;
}

.state-idle {
  background: rgba(107, 114, 128, 0.12);
  border: 1px solid rgba(107, 114, 128, 0.25);
  color: #6b7280;
}
.state-preparing, .state-measuring {
  background: rgba(59, 130, 246, 0.12);
  border: 1px solid rgba(59, 130, 246, 0.25);
  color: #2563eb;
}
.state-paused {
  background: rgba(245, 158, 11, 0.12);
  border: 1px solid rgba(245, 158, 11, 0.25);
  color: #d97706;
}
.state-completed {
  background: rgba(16, 185, 129, 0.12);
  border: 1px solid rgba(16, 185, 129, 0.25);
  color: #059669;
}
.state-error {
  background: rgba(239, 68, 68, 0.12);
  border: 1px solid rgba(239, 68, 68, 0.25);
  color: #dc2626;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 24px;
}

.status-info {
  display: flex;
  align-items: center;
  gap: 6px;
}

.info-label {
  color: #9ca3af;
  font-size: 12px;
  font-weight: 500;
  letter-spacing: 0.05em;
}

.info-value {
  font-size: 13px;
  font-weight: 600;

  &.stable { color: #22c55e; }
  &.unstable { color: #ef4444; }
}

.time-badge {
  font-family: 'JetBrains Mono', 'Fira Code', 'SF Mono', Consolas, monospace;
  background: #f3f4f6;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 13px;
  font-weight: 500;
  color: #374151;

  small {
    font-size: 10px;
    margin-left: 2px;
    color: #9ca3af;
  }
}

.workbench-content {
  flex: 1;
  min-height: 0;
  display: flex;
  gap: 16px;
  overflow: hidden;
}

.workbench-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
  overflow-y: auto;
  padding-right: 4px;
}

@media (max-width: 768px) {
  .page-header { padding: 0 16px; }
  .header-right { display: none; }
  .workbench-content { flex-direction: column; }
}
</style>
