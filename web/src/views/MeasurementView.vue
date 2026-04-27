<template>
  <PageLayout>
    <header class="page-header">
      <div class="header-left">
        <button class="back-btn" @click="goBack">
          <el-icon><ArrowLeft /></el-icon>
        </button>
        <div class="header-title">
          <h1>计量工作台</h1>
          <p>设备计量数据采集与管理</p>
        </div>
      </div>
      <div class="header-actions">
        <span class="state-badge" :class="stateClass">{{ stateLabel }}</span>
      </div>
    </header>

    <div class="workbench-content">
      <MeasurementSidebar
        ref="sidebarRef"
        :collapsed="sidebarCollapsed"
        @toggle="sidebarCollapsed = !sidebarCollapsed"
        @channels-change="handleChannelsChange"
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
function handleChannelsChange(channels: number[]) { console.log('Channels:', channels) }

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
  align-items: flex-end;
  flex-shrink: 0;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border-color-light);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.back-btn {
  width: 40px;
  height: 40px;
  background: rgba(30, 30, 30, 0.6);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  color: var(--text-secondary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;

  &:hover {
    background: rgba(50, 50, 50, 0.8);
    color: var(--text-primary);
    border-color: var(--border-color-strong);
  }
}

.header-title {
  h1 {
    font-size: 28px;
    font-weight: 700;
    color: var(--text-primary);
    margin: 0 0 4px;
    letter-spacing: -0.02em;
  }
  p {
    font-size: 13px;
    color: var(--text-secondary);
    margin: 0;
  }
}

.header-actions {
  display: flex;
  gap: 12px;
}

.state-badge {
  padding: 8px 16px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
}

.state-idle { background: rgba(100,100,100,0.3); color: var(--text-secondary); }
.state-preparing, .state-measuring { background: rgba(64,158,255,0.2); color: #409eff; }
.state-paused { background: rgba(230,162,60,0.2); color: #e6a23c; }
.state-completed { background: rgba(103,194,58,0.2); color: #67c23a; }
.state-error { background: rgba(245,108,108,0.2); color: #f56c6c; }

.workbench-content {
  flex: 1;
  min-height: 0;
  display: flex;
  gap: 24px;
  overflow: hidden;
}

.workbench-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 24px;
  overflow-y: auto;
}

@media (max-width: 768px) {
  .page-header { flex-direction: column; align-items: flex-start; gap: 16px; }
  .header-left { width: 100%; }
  .header-title h1 { font-size: 20px; }
  .workbench-content { flex-direction: column; }
}
</style>
