<template>
  <section class="control-bar secondary">
    <div class="param-group">
      <span class="param-label">采样状态</span>
      <span class="state-text">{{ stateText }}</span>
    </div>

    <div class="param-group">
      <span class="param-label">已选通道</span>
      <span class="channel-count">{{ selectedChannelCount }}/16</span>
    </div>

    <div class="param-group">
      <span class="param-label">稳定</span>
      <span
        class="status-badge"
        :class="isStable ? 'stable' : 'unstable'"
      >{{ isStable ? '是' : '否' }}</span>
    </div>

    <div class="param-group">
      <span class="param-label">时间</span>
      <span
        class="time-value"
        :class="{ stable: isStable }"
      >{{ stableSeconds.toFixed(1) }}s</span>
    </div>

    <div class="flex-spacer" />

    <div class="control-buttons">
      <button
        type="button"
        class="ctrl-btn btn-start"
        :disabled="!canStart"
        @click="$emit('start')"
      >
        <el-icon><VideoPlay /></el-icon>
        开始采样
      </button>
      <button
        type="button"
        class="ctrl-btn btn-pause"
        :disabled="!measurementStore.isRunning"
        @click="$emit('pause')"
      >
        <el-icon><VideoPause /></el-icon>
        暂停
      </button>
      <button
        type="button"
        class="ctrl-btn btn-resume"
        :disabled="!measurementStore.isPaused"
        @click="$emit('resume')"
      >
        恢复
      </button>
      <button
        type="button"
        class="ctrl-btn btn-stop"
        :disabled="measurementStore.isIdle"
        @click="$emit('stop')"
      >
        <el-icon><CloseBold /></el-icon>
        停止
      </button>
      <button
        v-if="measurementStore.totalRows > 0"
        type="button"
        class="ctrl-btn btn-export"
        @click="$emit('export')"
      >
        <el-icon><Download /></el-icon>
        导出CSV
      </button>
      <span
        v-if="measurementStore.totalRows > 0"
        class="row-count"
      >{{ measurementStore.totalRows }} 条采样</span>
      <span
        v-if="measurementStore.isRunning"
        class="realtime-pressure"
      >
        压力 {{ measurementStore.currentPressure?.toFixed(2) || '--' }} kPa
      </span>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, type PropType } from 'vue'
import { VideoPlay, VideoPause, CloseBold, Download } from '@element-plus/icons-vue'
import { useMeasurementStore } from '@/stores/measurement'

const props = defineProps({
  channels: {
    type: Array as PropType<number[]>,
    required: true
  },
  canStart: {
    type: Boolean,
    default: undefined
  },
  isStable: {
    type: Boolean,
    default: false
  },
  stableSeconds: {
    type: Number,
    default: 0
  },
  selectedChannelCount: {
    type: Number,
    default: 0
  }
})

defineEmits<{
  start: []
  pause: []
  resume: []
  stop: []
  export: []
}>()

const measurementStore = useMeasurementStore()

const canStart = computed(() => {
  if (typeof props.canStart === 'boolean') {
    return props.canStart
  }

  return measurementStore.isStartable && measurementStore.deviceBound
})

const stateTextMap: Record<string, string> = {
  idle: '空闲',
  pressuring: '打压中',
  stabilizing: '稳定中',
  collecting: '采集中',
  completed: '已完成',
  error: '错误',
  paused: '已暂停'
}

const stateText = computed(() => stateTextMap[measurementStore.state] || measurementStore.state)
</script>

<style scoped lang="scss">
.control-bar {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-sm) var(--spacing-md);
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  flex-wrap: wrap;
}

.param-group {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
}

.param-label {
  color: var(--text-muted);
  font-size: 11px;
  white-space: nowrap;
}

.segment-group {
  display: inline-flex;
  border: 1px solid var(--border-color-strong);
  border-radius: var(--radius-sm);
  overflow: hidden;
}

.segment-btn {
  height: 26px;
  min-width: 44px;
  padding: 0 10px;
  border: none;
  border-right: 1px solid var(--border-color-strong);
  background: var(--bg-secondary);
  color: var(--text-secondary);
  cursor: pointer;
  font-size: 12px;

  &:last-child {
    border-right: none;
  }

  &.active {
    background: var(--accent-primary);
    color: var(--bg-primary);
    font-weight: 600;
  }
}

.divider {
  width: 1px;
  height: 24px;
  background: var(--border-color-strong);
}

.progress-group {
  min-width: 180px;
}

.progress-container {
  min-width: 160px;
}

.progress-text {
  display: flex;
  justify-content: space-between;
  color: var(--text-secondary);
  font-size: 12px;
  margin-bottom: 3px;
}

.progress-percent {
  color: var(--accent-primary);
  font-weight: 600;
}

.progress-bar-bg {
  width: 100%;
  height: 6px;
  border-radius: 999px;
  background: var(--bg-quaternary);
  overflow: hidden;
}

.progress-bar-fill {
  height: 100%;
  background: var(--accent-primary);
  transition: width 0.2s ease;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 28px;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;

  &.stable {
    background: var(--status-success-bg);
    color: var(--status-success);
  }

  &.unstable {
    background: var(--status-warning-bg);
    color: var(--status-warning);
  }
}

.time-value {
  color: var(--text-secondary);
  font-variant-numeric: tabular-nums;
  font-size: 12px;

  &.stable {
    color: var(--status-success);
  }
}

.flex-spacer {
  flex: 1;
}

.alarm-group {
  gap: 6px;
}

.inline-check {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: var(--text-secondary);
  font-size: 12px;
  white-space: nowrap;

  input {
    margin: 0;
    accent-color: var(--accent-primary);
  }
}

.channel-select-btn {
  height: 26px;
  padding: 0 10px;
  border: 1px solid var(--border-color-strong);
  border-radius: var(--radius-sm);
  background: var(--bg-secondary);
  color: var(--text-secondary);
  font-size: 12px;
  cursor: default;
}

.channel-count {
  margin-left: 4px;
  color: var(--text-muted);
}

.control-buttons {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  flex-wrap: wrap;
}

.ctrl-btn {
  height: 34px;
  padding: 0 12px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color-strong);
  background: var(--bg-secondary);
  color: var(--text-secondary);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;

  &:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
}

.btn-start {
  background: color-mix(in srgb, var(--status-success) 22%, var(--bg-secondary));
  color: var(--status-success);
  border-color: color-mix(in srgb, var(--status-success) 45%, var(--border-color-strong));
}

.btn-stop {
  background: color-mix(in srgb, var(--status-error) 16%, var(--bg-secondary));
  color: var(--status-error);
  border-color: color-mix(in srgb, var(--status-error) 45%, var(--border-color-strong));
}

.btn-export {
  background: color-mix(in srgb, var(--status-info) 15%, var(--bg-secondary));
  color: var(--status-info);
}

.state-text {
  color: var(--text-muted);
  font-size: 12px;
}

.row-count,
.realtime-pressure {
  color: var(--text-secondary);
  font-size: 12px;
}

.realtime-pressure {
  font-variant-numeric: tabular-nums;
}

@media (max-width: 900px) {
  .control-bar {
    align-items: flex-start;
  }

  .divider,
  .flex-spacer {
    display: none;
  }

  .progress-group {
    width: 100%;
  }

  .control-buttons {
    width: 100%;
  }
}
</style>
