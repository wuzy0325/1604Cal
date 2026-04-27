<template>
  <section class="control-bar secondary">
    <div class="param-group">
      <span class="param-label">控制模式</span>
      <div class="segment-group">
        <button
          type="button"
          class="segment-btn"
          :class="{ active: measurementStore.measurementParams.controlMode === 'auto' }"
          @click="setControlMode('auto')"
        >
          自动
        </button>
        <button
          type="button"
          class="segment-btn"
          :class="{ active: measurementStore.measurementParams.controlMode === 'manual' }"
          @click="setControlMode('manual')"
        >
          手动
        </button>
      </div>
    </div>

    <div class="param-group">
      <span class="param-label">打压模式</span>
      <div class="segment-group">
        <button
          type="button"
          class="segment-btn"
          :class="{ active: measurementStore.measurementParams.pressureMode === 'single' }"
          @click="measurementStore.measurementParams.pressureMode = 'single'"
        >
          单程
        </button>
        <button
          type="button"
          class="segment-btn"
          :class="{ active: measurementStore.measurementParams.pressureMode === 'roundTrip' }"
          @click="measurementStore.measurementParams.pressureMode = 'roundTrip'"
        >
          回程
        </button>
      </div>
    </div>

    <div class="param-group">
      <span class="param-label">进度</span>
      <div class="progress-container">
        <div class="progress-text">
          <span class="progress-count">{{ completedCount }}/{{ totalCount }}</span>
          <span class="progress-percent">{{ progressPercent }}%</span>
        </div>
        <div class="progress-bar-bg">
          <div class="progress-bar-fill" :style="{ width: progressPercent + '%' }" />
        </div>
      </div>
    </div>

    <div class="divider" />

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

    <div class="param-group channel-group">
      <span class="param-label">采集通道</span>
      <button class="channel-select-btn" @click="channelDialogVisible = true">
        <el-icon><Grid /></el-icon>
        {{ measurementStore.channels.length }}/16
      </button>
    </div>

    <div class="divider" />

    <div class="param-group alarm-group">
      <span class="param-label">报警设置</span>
      <label class="inline-check">
        <input v-model="measurementStore.alarmConfig.enabled" type="checkbox" />
        <span>启用</span>
      </label>
      <label class="inline-check">
        <input v-model="measurementStore.alarmConfig.soundEnabled" type="checkbox" />
        <span>声音</span>
      </label>
      <label class="inline-check">
        <input v-model="measurementStore.alarmConfig.confirmOnAlarm" type="checkbox" />
        <span>报警确认</span>
      </label>
      <button class="channel-select-btn" @click="$emit('select-channel')">
        通道选择
        <span class="channel-count">({{ enabledChannelsDesc }})</span>
      </button>
    </div>

    <div class="flex-spacer" />

    <div class="control-buttons">
      <template v-if="measurementStore.measurementParams.controlMode === 'auto'">
        <button
          type="button"
          class="ctrl-btn btn-start btn-primary-action"
          :disabled="!canStart"
          @click="$emit('start')"
        >
          <el-icon><VideoPlay /></el-icon>
          开始采集
        </button>
      </template>
      <template v-else>
        <button
          type="button"
          class="ctrl-btn btn-start btn-primary-action"
          :disabled="!canManualPressurize"
          @click="$emit('manual-pressurize')"
        >
          手动打压
        </button>
        <button
          type="button"
          class="ctrl-btn btn-collect"
          :disabled="!canManualCollect"
          @click="$emit('manual-collect')"
        >
          采集
        </button>
      </template>
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
        class="ctrl-btn btn-stop btn-danger-action"
        :disabled="measurementStore.isIdle"
        @click="$emit('stop')"
      >
        <el-icon><CloseBold /></el-icon>
        停止
      </button>
      <button
        v-if="hasCompletedPoints"
        type="button"
        class="ctrl-btn btn-reset btn-danger-action"
        @click="$emit('reset')"
      >
        重置
      </button>
      <button
        v-if="measurementStore.totalRows > 0"
        type="button"
        class="ctrl-btn btn-export"
        @click="$emit('export')"
      >
        <el-icon><Download /></el-icon>
        导出报告
      </button>
    </div>

    <ChannelSelectDialog
      v-model:visible="channelDialogVisible"
      :selected-channels="measurementStore.channels"
      @confirm="handleChannelConfirm"
    />
  </section>
</template>

<script setup lang="ts">
import { computed, ref, type PropType } from 'vue'
import { VideoPlay, VideoPause, CloseBold, Download, Grid } from '@element-plus/icons-vue'
import { useMeasurementStore } from '@/stores/measurement'
import ChannelSelectDialog from '@/components/common/ChannelSelectDialog.vue'

const props = defineProps({
  channels: {
    type: Array as PropType<number[]>,
    default: undefined
  },
  canStart: {
    type: Boolean,
    default: undefined
  },
  isStable: {
    type: Boolean,
    default: undefined
  },
  stableSeconds: {
    type: Number,
    default: undefined
  },
  selectedChannelCount: {
    type: Number,
    default: undefined
  }
})

defineEmits<{
  start: []
  pause: []
  resume: []
  stop: []
  reset: []
  export: []
  'select-channel': []
  'manual-pressurize': []
  'manual-collect': []
}>()

const measurementStore = useMeasurementStore()

const channelDialogVisible = ref(false)

const handleChannelConfirm = (channels: number[]) => {
  measurementStore.channels = channels
}

const completedCount = computed(() =>
  measurementStore.currentPointIndex
)

const totalCount = computed(() => measurementStore.points.length)

const progressPercent = computed(() => {
  if (totalCount.value === 0) return 0
  return Math.round((completedCount.value / totalCount.value) * 100)
})

function setControlMode(mode: 'auto' | 'manual') {
  measurementStore.measurementParams.controlMode = mode
}

const hasCompletedPoints = computed(() =>
  measurementStore.points.some(p => p.status === 'completed')
)

const canManualPressurize = computed(() =>
  measurementStore.points.length > 0 && !measurementStore.isRunning
)

const canManualCollect = computed(() =>
  measurementStore.measurementParams.controlMode === 'manual' &&
  measurementStore.state === 'stabilizing' &&
  measurementStore.isStable
)

const enabledChannelsDesc = computed(() => {
  const chs = measurementStore.channels
  if (chs.length === 0) return '全部'
  if (chs.length <= 3) return chs.join(',')
  return `${chs[0]}-${chs[chs.length - 1]}`
})

const canStart = computed(() => {
  if (typeof props.canStart === 'boolean') {
    return props.canStart
  }

  return measurementStore.isStartable && measurementStore.deviceBound
})

const isStable = computed(() => {
  if (typeof props.isStable === 'boolean') {
    return props.isStable
  }
  return measurementStore.isStable
})

const stableSeconds = computed(() => {
  if (typeof props.stableSeconds === 'number') {
    return props.stableSeconds
  }
  return measurementStore.stabilityState.stableDurationMs / 1000
})
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

.progress-count {
  color: var(--text-secondary);
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
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 4px;

  &:hover {
    border-color: var(--accent-primary);
    color: var(--accent-primary);
  }
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

.btn-primary-action {
  background: var(--status-success) !important;
  color: var(--bg-primary) !important;
  border-color: var(--status-success) !important;
  font-weight: 600;
}

.btn-danger-action {
  background: color-mix(in srgb, var(--status-error) 16%, var(--bg-secondary));
  color: var(--status-error);
  border-color: color-mix(in srgb, var(--status-error) 45%, var(--border-color-strong));
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
