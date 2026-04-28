<template>
  <section class="control-card">
    <div class="control-row-top">
      <div class="mode-group">
        <div class="mode-item">
          <span class="mode-label">控制模式</span>
          <div class="segment-control">
            <button
              type="button"
              class="segment-btn"
              :class="{ active: controlMode === 'auto' }"
              @click="controlMode = 'auto'"
            >自动</button>
            <button
              type="button"
              class="segment-btn"
              :class="{ active: controlMode === 'manual' }"
              @click="controlMode = 'manual'"
            >手动</button>
          </div>
        </div>
      </div>

      <div class="channel-item">
        <span class="mode-label">采集通道</span>
        <button class="channel-select-btn" @click="channelDialogVisible = true">
          <el-icon><Grid /></el-icon>
          <span>{{ calibrationStore.selectedChannels.length }}/16</span>
        </button>
      </div>

      <div v-if="calibrationStore.pressurePoints.length > 0" class="progress-group">
        <div class="progress-labels">
          <span class="progress-text">进度 {{ completedCount }}/{{ calibrationStore.pressurePoints.length }}</span>
        </div>
        <div class="progress-track">
          <div class="progress-fill" :style="{ width: progressPercent + '%' }" />
        </div>
        <div class="stable-status">
          <span>{{ calibrationStore.isStable ? '已稳定' : '稳定中' }}</span>
          <span class="session-state">会话: {{ sessionStateText }}</span>
        </div>
      </div>

      <div v-else class="session-status-inline">
        <span :class="['status-badge', `status-${sessionState}`]">
          {{ sessionStateText }}
        </span>
        <span v-if="!valveReady" class="valve-warning">
          阀门未切换到校准状态
        </span>
      </div>
    </div>

    <div v-if="controlMode === 'manual' && isRunning" class="manual-controls">
      <ManualControlPanel
        :max-pressure="calibrationParams.maxValue"
        :stability-status="stabilityStatus"
        @collected="handleManualCollect"
      />
    </div>

    <div v-else class="control-row-bottom">
      <template v-if="sessionState === 'await_alarm_resolution'">
        <button
          type="button"
          class="ctrl-btn btn-warning"
          @click="calibrationStore.resolveAlarm('continue')"
        >
          <el-icon><CircleCheck /></el-icon>
          报警确认继续
        </button>
        <button
          type="button"
          class="ctrl-btn btn-stop"
          @click="calibrationStore.resolveAlarm('recollect')"
        >
          <el-icon><RefreshRight /></el-icon>
          报警重采
        </button>
      </template>
      <template v-else>
        <button
          type="button"
          class="ctrl-btn btn-start"
          :disabled="calibrationStore.isRunning"
          @click="calibrationStore.startCalibration()"
        >
          <el-icon><VideoPlay /></el-icon>
          开始
        </button>
        <button
          type="button"
          class="ctrl-btn btn-pause"
          :disabled="!calibrationStore.isRunning"
          @click="calibrationStore.pauseCalibration()"
        >
          <el-icon><VideoPause /></el-icon>
          暂停
        </button>
        <button
          type="button"
          class="ctrl-btn btn-resume"
          :disabled="sessionState !== 'paused'"
          @click="calibrationStore.resumeCalibration()"
        >
          <el-icon><RefreshRight /></el-icon>
          继续
        </button>
        <button
          type="button"
          class="ctrl-btn btn-stop"
          :disabled="sessionState === 'idle' || sessionState === 'stopped'"
          @click="calibrationStore.stopCalibration()"
        >
          <el-icon><CloseBold /></el-icon>
          停止
        </button>
        <div class="action-divider" />
        <button
          type="button"
          class="ctrl-btn btn-fit"
          :disabled="!calibrationStore.hasCollectedData"
          @click="calibrationStore.fitData()"
        >
          <el-icon><DataAnalysis /></el-icon>
          拟合
        </button>
        <button
          type="button"
          class="ctrl-btn btn-end"
          @click="calibrationStore.endCalibration()"
        >
          <el-icon><CircleClose /></el-icon>
          结束
        </button>
      </template>
    </div>

    <ChannelSelectDialog
      v-model:visible="channelDialogVisible"
      :selected-channels="calibrationStore.selectedChannels"
      @confirm="calibrationStore.setSelectedChannels"
    />
  </section>
</template>

<script setup lang="ts">
import { computed, inject, ref } from 'vue'
import {
  VideoPlay,
  VideoPause,
  RefreshRight,
  CloseBold,
  DataAnalysis,
  CircleClose,
  CircleCheck,
  Grid
} from '@element-plus/icons-vue'
import type { SessionState } from '@/types/calibration'
import { useCalibrationStore } from '@/stores/calibration'
import ManualControlPanel from './ManualControlPanel.vue'
import ChannelSelectDialog from '@/components/common/ChannelSelectDialog.vue'
import { stabilityStatusKey } from '@/composables/useCalibrationSync'

const calibrationStore = useCalibrationStore()
const stabilityStatus = inject(stabilityStatusKey)!
const channelDialogVisible = ref(false)

const controlMode = computed<'auto' | 'manual'>({
  get: () => calibrationStore.controlMode,
  set: (mode) => {
    calibrationStore.controlMode = mode
  }
})
const sessionState = computed(() => calibrationStore.sessionState)
const isRunning = computed(() => calibrationStore.isRunning)
const valveStatus = computed(() => calibrationStore.valveStatus)
const valveReady = computed(() => valveStatus.value === 'calibration')

const sessionStateTextMap: Record<SessionState, string> = {
  idle: '空闲',
  ready: '就绪',
  pressurizing: '打压中',
  stabilizing: '稳定中',
  collecting: '采集中',
  point_done: '点完成',
  fitting: '拟合中',
  completed: '已完成',
  paused: '已暂停',
  stopped: '已停止',
  await_manual_collect: '等待手动采集',
  await_alarm_resolution: '等待报警处理',
  recovering: '恢复中',
  error: '错误'
}

const sessionStateText = computed(() => sessionStateTextMap[sessionState.value] || sessionState.value)

const completedCount = computed(() =>
  calibrationStore.pressurePoints.filter(p => p.status === 'completed').length
)
const progressPercent = computed(() => {
  const total = calibrationStore.pressurePoints.length
  if (total === 0) return 0
  return Math.round((completedCount.value / total) * 100)
})

const calibrationParams = computed(() => calibrationStore.calibrationParams)

function handleManualCollect(data: number[]) {
  console.log('Manual collection complete:', data.length, 'channels')
}
</script>

<style scoped lang="scss">
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
}

.channel-item {
  display: flex;
  align-items: center;
  gap: 8px;
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

.stable-status {
  display: flex;
  gap: 16px;
  color: $slate-500;
  font-size: 11px;

  .session-state {
    color: $mint;
    font-weight: 600;
  }
}

.session-status-inline {
  display: flex;
  align-items: center;
  gap: 10px;
}

.valve-warning {
  color: #d97706;
  font-size: 11px;
  font-weight: 500;
  background: rgba(245, 158, 11, 0.1);
  border: 1px solid rgba(245, 158, 11, 0.2);
  padding: 2px 8px;
  border-radius: 4px;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 3px 10px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  line-height: 1.5;
  letter-spacing: 0.02em;
}

.status-idle { background: rgba(107, 114, 128, 0.12); border: 1px solid rgba(107, 114, 128, 0.25); color: $slate-500; }
.status-ready { background: rgba(59, 130, 246, 0.12); border: 1px solid rgba(59, 130, 246, 0.25); color: #2563eb; }

.status-pressurizing,
.status-stabilizing,
.status-collecting,
.status-point_done,
.status-fitting,
.status-await_manual_collect,
.status-await_alarm_resolution,
.status-recovering { background: rgba(16, 185, 129, 0.12); border: 1px solid rgba(16, 185, 129, 0.25); color: #059669; }

.status-paused { background: rgba(245, 158, 11, 0.12); border: 1px solid rgba(245, 158, 11, 0.25); color: #d97706; }
.status-completed { background: rgba(16, 185, 129, 0.12); border: 1px solid rgba(16, 185, 129, 0.25); color: #059669; }
.status-stopped { background: rgba(107, 114, 128, 0.12); border: 1px solid rgba(107, 114, 128, 0.25); color: $slate-500; }
.status-error { background: rgba(239, 68, 68, 0.12); border: 1px solid rgba(239, 68, 68, 0.25); color: #dc2626; }

.control-row-bottom {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

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

.btn-stop {
  background: rgba(239, 68, 68, 0.1);
  color: $red;
  border: 1px solid rgba(239, 68, 68, 0.2);

  &:hover:not(:disabled) {
    background: rgba(239, 68, 68, 0.16);
    border-color: rgba(239, 68, 68, 0.35);
  }
}

.btn-fit {
  background: rgba(59, 130, 246, 0.1);
  color: $blue;
  border: 1px solid rgba(59, 130, 246, 0.2);

  &:hover:not(:disabled) {
    background: rgba(59, 130, 246, 0.16);
    border-color: rgba(59, 130, 246, 0.35);
  }
}

.btn-end {
  background: rgba(55, 65, 81, 0.08);
  color: $slate-700;
  border: 1px solid $slate-200;

  &:hover:not(:disabled) {
    background: rgba(55, 65, 81, 0.14);
    border-color: $slate-300;
  }
}

.btn-warning {
  background: rgba(245, 158, 11, 0.1);
  color: $amber;
  border: 1px solid rgba(245, 158, 11, 0.2);

  &:hover:not(:disabled) {
    background: rgba(245, 158, 11, 0.16);
    border-color: rgba(245, 158, 11, 0.35);
  }
}

.action-divider {
  width: 1px;
  height: 20px;
  background: $slate-200;
  margin: 0 4px;
}

.manual-controls {
  display: flex;
  flex-direction: column;
  min-width: 240px;
}

@media (max-width: 1200px) {
  .mode-group {
    gap: 16px;
  }

  .control-row-top {
    gap: 12px;
  }
}

@media (max-width: 900px) {
  .control-row-top {
    flex-direction: column;
    align-items: flex-start;
  }

  .progress-group {
    width: 100%;
  }

  .control-row-bottom {
    width: 100%;
  }
}
</style>
