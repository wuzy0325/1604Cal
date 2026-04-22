<template>
  <div class="control-panel">
    <div class="control-panel-top">
      <div class="param-group">
        <div class="control-group">
          <label>采集通道</label>
          <el-input-number :model-value="channels.length" :min="1" :max="16" disabled size="small" />
        </div>
      </div>
      <div class="param-divider" />
      <div class="status-section">
        <div class="measurement-status">
          <span :class="['status-badge', `status-${measurementStore.state}`]">{{ stateText }}</span>
          <span v-if="measurementStore.isCollecting" class="row-count">已采集 {{ measurementStore.totalRows }} 行</span>
          <span v-if="measurementStore.isCollecting" class="realtime-pressure">压力: {{ measurementStore.currentPressure?.toFixed(2) || '--' }} kPa</span>
        </div>
      </div>
    </div>
    <div class="control-panel-bottom">
      <div class="action-buttons">
        <el-button type="success" :disabled="!canStart" @click="$emit('start')">
          <el-icon><VideoPlay /></el-icon>开始采集
        </el-button>
        <el-button :disabled="!measurementStore.isCollecting" @click="$emit('pause')">
          <el-icon><VideoPause /></el-icon>暂停
        </el-button>
        <el-button :disabled="!measurementStore.isPaused" @click="$emit('resume')">
          <el-icon><RefreshRight /></el-icon>恢复
        </el-button>
        <el-button type="danger" :disabled="measurementStore.isIdle" @click="$emit('stop')">
          <el-icon><CloseBold /></el-icon>停止
        </el-button>
        <el-button v-if="measurementStore.totalRows > 0" type="primary" @click="$emit('export')">
          <el-icon><Download /></el-icon>导出CSV
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { VideoPlay, VideoPause, RefreshRight, CloseBold, Download } from '@element-plus/icons-vue'
import { useMeasurementStore } from '@/stores/measurement'

defineProps<{ channels: number[] }>()
defineEmits<{ start: []; pause: []; resume: []; stop: []; export: [] }>()

const measurementStore = useMeasurementStore()

const canStart = computed(() => measurementStore.isIdle && measurementStore.deviceBound)

const stateTextMap: Record<string, string> = {
  idle: '空闲', collecting: '采集中', paused: '已暂停'
}
const stateText = computed(() => stateTextMap[measurementStore.state] || measurementStore.state)
</script>

<style scoped lang="scss">
.control-panel { background: var(--bg-secondary); border: 1px solid var(--border-color); border-radius: var(--radius-md); display: flex; flex-direction: column; flex-shrink: 0; overflow: hidden; }
.control-panel-top { padding: var(--spacing-md) var(--spacing-lg); display: flex; align-items: center; gap: var(--spacing-xl); border-bottom: 1px solid var(--border-color); }
.param-group { display: flex; gap: var(--spacing-md); align-items: flex-end; }
.param-divider { width: 1px; height: 32px; background: var(--border-color-strong); flex-shrink: 0; align-self: center; }
.control-group { display: flex; flex-direction: column; gap: 4px; label { color: var(--text-muted); font-size: 11px; font-weight: 500; white-space: nowrap; } :deep(.el-input-number) { width: 90px; } }
.status-section { flex: 1; }
.measurement-status { display: flex; gap: var(--spacing-md); align-items: center; font-size: 13px; }
.status-badge { display: inline-flex; align-items: center; padding: 4px 10px; border-radius: var(--radius-sm); font-size: 12px; font-weight: 600; }
.status-idle { background: var(--bg-quaternary); color: var(--text-secondary); }
.status-collecting { background: var(--status-success-bg); color: var(--status-success); }
.status-paused { background: var(--status-warning-bg); color: var(--status-warning); }
.row-count { color: var(--text-secondary); font-size: 12px; }
.realtime-pressure { color: var(--text-primary); font-weight: 600; font-variant-numeric: tabular-nums; font-size: 13px; }
.control-panel-bottom { padding: var(--spacing-sm) var(--spacing-lg); background: var(--bg-tertiary); display: flex; align-items: center; gap: var(--spacing-md); }
.action-buttons { display: flex; gap: var(--spacing-xs); flex-wrap: wrap; align-items: center; }
@media (max-width: 900px) {
  .control-panel-top { flex-direction: column; align-items: stretch; gap: var(--spacing-md); }
  .param-divider { display: none; }
  .measurement-status { flex-wrap: wrap; }
  .action-buttons { justify-content: center; flex-wrap: wrap; }
}
</style>
