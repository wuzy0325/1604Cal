<template>
  <section class="control-card">
    <div class="control-main">
      <div v-if="calibrationStore.pressurePoints.length > 0" class="progress-group">
        <div class="progress-labels">
          <span class="progress-text">进度 {{ completedCount }}/{{ calibrationStore.pressurePoints.length }}</span>
          <span class="stable-label">{{ calibrationStore.isStable ? '已稳定' : '稳定中' }}</span>
        </div>
        <div class="progress-track">
          <div class="progress-fill" :style="{ width: progressPercent + '%' }" />
        </div>
      </div>

      <div class="action-group">
        <template v-if="sessionState === 'await_alarm_resolution'">
          <button
            type="button"
            class="ctrl-btn btn-warning"
            @click="calibrationStore.resolveAlarm('continue')"
          >
            <el-icon><CircleCheck /></el-icon>
            确认继续
          </button>
          <button
            type="button"
            class="ctrl-btn btn-fit"
            @click="calibrationStore.resolveAlarm('skip')"
          >
            <el-icon><Right /></el-icon>
            跳过此点
          </button>
          <button
            type="button"
            class="ctrl-btn btn-resume"
            @click="calibrationStore.resolveAlarm('recollect')"
          >
            <el-icon><RefreshRight /></el-icon>
            重新采集
          </button>
          <button
            type="button"
            class="ctrl-btn btn-stop"
            @click="calibrationStore.resolveAlarm('stop')"
          >
            <el-icon><CloseBold /></el-icon>
            停止标定
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
            :disabled="!canPause"
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
            :disabled="!canFit"
            @click="calibrationStore.fitData()"
          >
            <el-icon><DataAnalysis /></el-icon>
            拟合
          </button>
          <button
            type="button"
            class="ctrl-btn btn-end"
            :disabled="!canEnd"
            @click="calibrationStore.endCalibration()"
          >
            <el-icon><CircleClose /></el-icon>
            结束
          </button>
        </template>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  VideoPlay,
  VideoPause,
  RefreshRight,
  CloseBold,
  DataAnalysis,
  CircleClose,
  CircleCheck,
  Right
} from '@element-plus/icons-vue'
import { useCalibrationStore } from '@/stores/calibration'

const calibrationStore = useCalibrationStore()

const sessionState = computed(() => calibrationStore.sessionState)
const isRunning = computed(() => calibrationStore.isRunning)

// 可暂停的状态：运行中的采集/打压/稳定阶段，排除拟合和报警等待
const canPause = computed(() =>
  isRunning.value &&
  sessionState.value !== 'fitting' &&
  sessionState.value !== 'await_alarm_resolution'
)

// 可拟合：有数据且未在拟合/已完成状态
const canFit = computed(() =>
  calibrationStore.hasCollectedData &&
  sessionState.value !== 'fitting' &&
  sessionState.value !== 'completed'
)

// 可结束：非空闲状态
const canEnd = computed(() =>
  sessionState.value !== 'idle' && sessionState.value !== 'stopped'
)

const completedCount = computed(() =>
  calibrationStore.pressurePoints.filter(p => p.status === 'completed').length
)
const progressPercent = computed(() => {
  const total = calibrationStore.pressurePoints.length
  if (total === 0) return 0
  return Math.round((completedCount.value / total) * 100)
})

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
  padding: 16px;
  display: flex;
  flex-direction: column;
  font-family: $font-sans;
  transition: box-shadow 0.2s ease;

  &:hover {
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.07), 0 2px 4px -2px rgba(0, 0, 0, 0.05);
  }
}

/* ── 进度 + 操作按钮 并列 ── */
.control-main {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  flex-wrap: wrap;
}

.progress-group {
  flex: 0 0 200px;
  display: flex;
  flex-direction: column;
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

.stable-label {
  font-size: 11px;
  color: $mint;
  font-weight: 600;
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

.action-group {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  gap: 8px;
  flex: 1;
  min-width: 280px;
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

@media (max-width: 900px) {
  .control-main {
    flex-direction: column;
    align-items: stretch;
  }

  .progress-group {
    width: 100%;
  }

  .action-group {
    width: 100%;
  }
}
</style>
