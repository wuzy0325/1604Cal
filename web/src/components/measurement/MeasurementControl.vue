<template>
  <section class="control-card">
    <!-- 第一排：模式与进度 -->
    <div class="control-row-top">
      <div class="mode-group">
        <div class="mode-item">
          <span class="mode-label">模式</span>
          <div class="segment-control">
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

        <div class="mode-item">
          <span class="mode-label">打压</span>
          <div class="segment-control">
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
              class="segment-btn pressure-active"
              :class="{ active: measurementStore.measurementParams.pressureMode === 'roundTrip' }"
              @click="measurementStore.measurementParams.pressureMode = 'roundTrip'"
            >
              回程
            </button>
          </div>
        </div>
      </div>

      <!-- 进度条 -->
      <div class="progress-group">
        <div class="progress-labels">
          <span class="progress-text">任务进度: {{ completedCount }}/{{ totalCount }}</span>
          <span class="progress-percent">{{ progressPercent }}%</span>
        </div>
        <div class="progress-track">
          <div class="progress-fill" :style="{ width: progressPercent + '%' }" />
        </div>
      </div>

      <!-- 操作按钮组 -->
      <div class="control-buttons">
        <template v-if="measurementStore.measurementParams.controlMode === 'auto'">
          <button
            type="button"
            class="ctrl-btn btn-start"
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
            class="ctrl-btn btn-start"
            :disabled="!canManualStart"
            @click="$emit('manual-start')"
          >
            开始
          </button>
          <button
            v-if="hasPressureDevice"
            type="button"
            class="ctrl-btn btn-start"
            :disabled="!canManualPressurize"
            @click="$emit('manual-pressurize')"
          >
            手动打压
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
          class="ctrl-btn btn-stop"
          :disabled="measurementStore.isIdle"
          @click="$emit('stop')"
        >
          <el-icon><CloseBold /></el-icon>
          停止
        </button>
        <button
          v-if="hasCompletedPoints"
          type="button"
          class="ctrl-btn btn-reset"
          @click="$emit('reset')"
        >
          重置
        </button>
        <button
          v-if="hasCompletedPoints"
          type="button"
          class="ctrl-btn btn-export"
          @click="$emit('export')"
        >
          <el-icon><Download /></el-icon>
          导出报告
        </button>

      </div>
    </div>

    <!-- 第二排：通道与报警 -->
    <div class="control-row-bottom">
      <div class="left-controls">
        <div class="channel-item">
          <span class="mode-label">采集通道</span>
          <button class="channel-select-btn" @click="channelDialogVisible = true">
            <el-icon><Grid /></el-icon>
            <span>{{ measurementStore.channels.length }}/16</span>
          </button>
        </div>

        <div class="alarm-item">
          <span class="mode-label">报警设置</span>
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
        </div>
      </div>
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
  },
  hasPressureDevice: {
    type: Boolean,
    default: true
  }
})

defineEmits<{
  start: []
  pause: []
  resume: []
  stop: []
  reset: []
  export: []
  'manual-start': []
  'manual-pressurize': []
}>()

const measurementStore = useMeasurementStore()

const channelDialogVisible = ref(false)

const handleChannelConfirm = (channels: number[]) => {
  measurementStore.channels = channels
}

const completedCount = computed(() =>
  measurementStore.points.filter(p => p.status === 'completed').length
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
  props.hasPressureDevice &&
  measurementStore.points.length > 0 &&
  !measurementStore.isRunning
)

const canManualStart = computed(() =>
  measurementStore.measurementParams.controlMode === 'manual' &&
  measurementStore.points.length > 0 &&
  ['idle', 'stopped', 'completed'].includes(measurementStore.state)
)

const canStart = computed(() => {
  if (typeof props.canStart === 'boolean') {
    return props.canStart
  }

  return measurementStore.isStartable && measurementStore.deviceBound
})

</script>

<style scoped lang="scss">
/* ── 设计系统令牌 ── */
$font-sans: 'Inter', 'SF Pro Display', -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif;
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
$green: #22c55e;
$red: #ef4444;
$blue: #3b82f6;
$amber: #f59e0b;

.control-card {
  background: #ffffff;
  border-radius: 12px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  border: 1px solid $slate-200;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  font-family: $font-sans;
  transition: box-shadow 0.2s ease, border-color 0.2s ease;

  &:hover {
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.07), 0 2px 4px -2px rgba(0, 0, 0, 0.05);
  }
}

.control-row-top {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 16px 24px;
}

.mode-group {
  display: flex;
  align-items: center;
  gap: 20px;
}

.mode-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

/* Label 规范：500, 12px, 1.5, 0.05em */
.mode-label {
  font-size: 12px;
  color: $slate-500;
  font-weight: 500;
  letter-spacing: 0.05em;
  white-space: nowrap;
  font-family: $font-sans;
}

.segment-control {
  display: flex;
  padding: 2px;
  background: $slate-100;
  border-radius: 8px;
  border: 1px solid $slate-200;
}

.segment-btn {
  padding: 4px 14px;
  font-size: 12px;
  font-weight: 500;
  border: none;
  background: transparent;
  color: $slate-500;
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.15s ease;
  font-family: $font-sans;

  &:hover {
    color: $slate-700;
  }

  &.active {
    background: linear-gradient(135deg, $mint, $mint-dark);
    color: #fff;
    box-shadow: 0 1px 3px rgba(16, 185, 129, 0.25);
  }

  &.pressure-active.active {
    background: linear-gradient(135deg, $blue, #2563eb);
    box-shadow: 0 1px 3px rgba(59, 130, 246, 0.25);
  }
}

.progress-group {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 160px;
  max-width: 400px;
  gap: 6px;
}

.progress-labels {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.progress-text {
  font-size: 12px;
  color: $slate-500;
  font-weight: 500;
  letter-spacing: 0.02em;
  font-family: $font-sans;
}

.progress-percent {
  font-size: 12px;
  font-weight: 700;
  color: $mint-dark;
  font-family: $font-sans;
}

.progress-track {
  width: 100%;
  height: 6px;
  background: $slate-100;
  border-radius: 999px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, $mint, $mint-light);
  transition: width 0.3s ease;
}

.control-buttons {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-left: auto;
}

/* 按钮基础：8px radius，规范过渡 */
.ctrl-btn {
  padding: 8px 16px;
  border-radius: 8px;
  border: none;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 600;
  transition: all 0.15s ease;
  font-family: $font-sans;

  &:active {
    transform: translateY(1px);
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

/* Primary：渐变 + 白色文字 + Mint 阴影 */
.btn-start {
  background: linear-gradient(135deg, $mint, $mint-dark);
  color: #fff;
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.3);

  &:hover:not(:disabled) {
    background: linear-gradient(135deg, $mint-light, $mint);
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.4);
    transform: translateY(-1px);
  }
}

/* Default：半透明 slate */
.btn-pause {
  background: rgba(55, 65, 81, 0.08);
  color: $slate-400;
  border: 1px solid $slate-200;

  &:hover:not(:disabled) {
    background: rgba(55, 65, 81, 0.14);
    color: $slate-500;
    border-color: $slate-300;
  }
}

.btn-resume {
  background: rgba(55, 65, 81, 0.08);
  color: $slate-700;
  border: 1px solid $slate-200;

  &:hover:not(:disabled) {
    background: rgba(55, 65, 81, 0.14);
    border-color: $slate-300;
  }
}

/* Stop Red */
.btn-stop {
  background: rgba(239, 68, 68, 0.1);
  color: $red;
  border: 1px solid rgba(239, 68, 68, 0.2);

  &:hover:not(:disabled) {
    background: rgba(239, 68, 68, 0.16);
    border-color: rgba(239, 68, 68, 0.35);
  }
}

.btn-reset {
  background: rgba(55, 65, 81, 0.08);
  color: $slate-700;
  border: 1px solid $slate-200;

  &:hover:not(:disabled) {
    background: rgba(55, 65, 81, 0.14);
    border-color: $slate-300;
  }
}

/* Info Blue */
.btn-export {
  background: rgba(59, 130, 246, 0.1);
  color: $blue;
  border: 1px solid rgba(59, 130, 246, 0.2);

  &:hover:not(:disabled) {
    background: rgba(59, 130, 246, 0.16);
    border-color: rgba(59, 130, 246, 0.35);
  }
}

.control-row-bottom {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 16px;
  padding-top: 12px;
  border-top: 1px solid $slate-100;
}

.left-controls {
  display: flex;
  align-items: center;
  gap: 20px;
  flex-wrap: wrap;
}

.channel-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.alarm-item {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.inline-check {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: $slate-600;
  font-size: 12px;
  cursor: pointer;
  white-space: nowrap;
  font-family: $font-sans;

  input[type="checkbox"] {
    width: 14px;
    height: 14px;
    accent-color: $mint;
    border: 1px solid $slate-300;
    border-radius: 3px;
    cursor: pointer;
  }

  &:hover span {
    color: $slate-800;
  }
}

.channel-select-btn {
  height: 28px;
  padding: 0 10px;
  border: 1px solid $slate-200;
  border-radius: 8px;
  background: $slate-50;
  color: $blue;
  font-size: 12px;
  font-family: $font-mono;
  font-weight: 500;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  transition: all 0.15s ease;

  &:hover {
    background: $slate-100;
    border-color: $slate-300;
  }

  .el-icon {
    font-size: 12px;
    color: $slate-400;
  }
}

.channel-count {
  margin-left: 2px;
  color: $slate-400;
  font-size: 11px;
}

@media (max-width: 900px) {
  .control-row-top {
    flex-direction: column;
    align-items: flex-start;
  }

  .progress-group {
    width: 100%;
  }

  .control-buttons {
    width: 100%;
    margin-left: 0;
  }

  .control-row-bottom {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
