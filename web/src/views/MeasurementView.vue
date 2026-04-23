<template>
  <section class="module-page">
    <div class="desktop-shell">
      <header class="module-header">
        <div>
          <p class="module-caption">Measurement Workbench</p>
          <h1>计量工作台</h1>
        </div>
        <nav class="module-switch">
          <RouterLink class="switch-btn" :to="{ name: 'module-device-management' }">设备管理</RouterLink>
          <RouterLink class="switch-btn active" :to="{ name: 'module-measurement' }">计量模块</RouterLink>
          <RouterLink class="switch-btn" :to="{ name: 'module-calibration' }">标定模块</RouterLink>
          <RouterLink class="switch-btn" :to="{ name: 'module-multi-pressure' }">多设备打压</RouterLink>
          <RouterLink class="switch-btn switch-btn-ghost" :to="{ name: 'module-hub' }">
            <el-icon><ArrowLeft /></el-icon>返回
          </RouterLink>
        </nav>
      </header>

      <div class="measurement-layout">
        <MeasurementSidebar
          ref="sidebarRef"
          :collapsed="sidebarCollapsed"
          @toggle="sidebarCollapsed = !sidebarCollapsed"
        />
        <main class="workbench">
          <MeasurementControl
            :channels="measurementStore.channels"
            @start="handleStart"
            @pause="handlePause"
            @resume="handleResume"
            @stop="handleStop"
            @export="exportCSV"
          />
          <MeasurementDataView
            :rows="measurementStore.rows"
            :channels="measurementStore.channels"
            @export-csv="exportCSV"
          />
        </main>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { RouterLink } from 'vue-router'
import { ArrowLeft } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

import MeasurementSidebar from '@/components/measurement/MeasurementSidebar.vue'
import MeasurementControl from '@/components/measurement/MeasurementControl.vue'
import MeasurementDataView from '@/components/measurement/MeasurementDataView.vue'
import { useMeasurementStore } from '@/stores/measurement'
import { useMeasurementDeviceStore } from '@/stores/measurement/deviceStore'

const measurementStore = useMeasurementStore()
const deviceStore = useMeasurementDeviceStore()

const sidebarCollapsed = ref(false)
const sidebarRef = ref<InstanceType<typeof MeasurementSidebar> | null>(null)

// ── 采集工作流 ──

const handleStart = () => {
  const channels = sidebarRef.value?.selectedChannels ?? []
  if (channels.length === 0) {
    ElMessage.warning('请先选择采集通道')
    return
  }
  measurementStore.start(channels)
}

const handlePause = () => measurementStore.pause()
const handleResume = () => {
  // 恢复时重新启动（后端 pause → idle，resume 需重新 start）
  const channels = measurementStore.channels
  if (channels.length > 0) {
    measurementStore.start(channels)
  }
}
const handleStop = () => measurementStore.stop()

// ── CSV 导出 ──

const exportCSV = () => {
  const rows = measurementStore.rows
  if (rows.length === 0) {
    ElMessage.warning('没有可导出的数据')
    return
  }
  const channels = measurementStore.channels
  const headers = ['时间', ...channels.map(ch => `CH${ch}`)]
  const csvRows = rows.map(r => [
    r.timestamp,
    ...channels.map(ch => r.channels[String(ch)]?.toFixed(4) ?? '')
  ])
  const csvContent = [headers.join(','), ...csvRows.map(r => r.join(','))].join('\n')
  const blob = new Blob(['\ufeff' + csvContent], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  link.href = URL.createObjectURL(blob)
  link.download = `measurement_${new Date().toISOString().split('T')[0]}.csv`
  link.click()
  ElMessage.success('CSV 已导出')
}

// ── 轮询 ──

let pollTimer: ReturnType<typeof setInterval> | null = null
let deviceRefreshTimer: ReturnType<typeof setInterval> | null = null

const startPolling = () => {
  if (pollTimer) return
  pollTimer = setInterval(async () => {
    if (measurementStore.isRunning) {
      await Promise.all([
        measurementStore.refreshPressure(),
        measurementStore.refreshStability()
      ])
    }
  }, 2000)
}

const startDeviceRefresh = () => {
  if (deviceRefreshTimer) return
  deviceRefreshTimer = setInterval(() => {
    deviceStore.loadDevices(true)
  }, 5000)
}

const stopPolling = () => {
  if (pollTimer) { clearInterval(pollTimer); pollTimer = null }
  if (deviceRefreshTimer) { clearInterval(deviceRefreshTimer); deviceRefreshTimer = null }
}

// ── 生命周期 ──

onMounted(async () => {
  await deviceStore.loadDevices()
  await measurementStore.fetchCurrentState()

  // 自动绑定已连接设备，保证进入页面即可读取会话数据。
  const connectedMeasureDev = deviceStore.measureDevices.find(d => d.status === 'connected')
  const connectedPressureDev = deviceStore.pressureDevices.find(d => d.status === 'connected')

  if (connectedMeasureDev && connectedPressureDev) {
    await measurementStore.bindDevices(connectedMeasureDev.id, connectedPressureDev.id)
  } else if (connectedMeasureDev) {
    await measurementStore.bindMeasureDevice(connectedMeasureDev.id)
  }

  if (connectedMeasureDev) {
    await Promise.all([
      measurementStore.refreshDeviceInfo(),
      measurementStore.refreshValveStatus(),
      measurementStore.refreshMeasureUnit()
    ])
  }

  sidebarRef.value?.checkUnitConsistency()

  measurementStore.setupSSE()
  startPolling()
  startDeviceRefresh()
})

onUnmounted(() => {
  measurementStore.teardownSSE()
  stopPolling()
})
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
.module-caption {
  color: var(--accent-primary);
  font-size: 11px;
  letter-spacing: 0.08em;
  margin: 0 0 var(--spacing-xs);
  text-transform: uppercase;
  font-weight: 600;
}
.module-header h1 {
  color: var(--text-primary);
  margin: 0;
  font-size: 20px;
  font-weight: 600;
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
}
.switch-btn.active {
  background: var(--accent-primary);
  border-color: var(--accent-primary);
  color: var(--bg-primary);
  font-weight: 600;
}
.switch-btn-ghost { background: transparent; }
.measurement-layout {
  flex: 1;
  min-height: 0;
  display: flex;
  gap: 0;
  background: var(--bg-primary);
}
.workbench {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
  padding: var(--spacing-md) var(--spacing-lg);
  overflow-y: auto;
  overflow-x: hidden;
}
@media (max-width: 900px) {
  .module-page { padding: var(--spacing-md); overflow: auto; }
  .desktop-shell { max-width: 100%; height: auto; gap: var(--spacing-md); }
  .module-header { flex-direction: column; gap: var(--spacing-md); }
  .measurement-layout { flex-direction: column; min-height: auto; }
  .workbench { overflow: visible; padding: var(--spacing-md); }
}
</style>
