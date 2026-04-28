<template>
  <section class="workbench-section control-section">
    <div class="control-row">
      <div class="mode-switches">
        <div class="switch-group">
          <span>控制模式</span>
          <el-radio-group v-model="controlMode" size="small">
            <el-radio-button label="auto">自动</el-radio-button>
            <el-radio-button label="manual">手动</el-radio-button>
          </el-radio-group>
        </div>
      </div>

      <div class="channel-select-group">
        <span class="channel-label">采集通道</span>
        <el-button size="small" @click="channelDialogVisible = true">
          <el-icon><Grid /></el-icon>
          {{ calibrationStore.selectedChannels.length }}/16
        </el-button>
      </div>

      <div
        v-if="calibrationStore.pressurePoints.length > 0"
        class="progress-section"
      >
        <div class="progress-info">
          <span>进度 {{ completedCount }}/{{ calibrationStore.pressurePoints.length }}</span>
          <el-progress :percentage="progressPercent" :stroke-width="8" />
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

      <div v-if="controlMode === 'manual' && isRunning" class="manual-controls">
        <ManualControlPanel
          :max-pressure="calibrationParams.maxValue"
          :stability-status="stabilityStatus"
          @collected="handleManualCollect"
        />
      </div>

      <div v-else class="action-buttons">
        <template v-if="sessionState === 'await_alarm_resolution'">
          <el-button
            type="warning"
            @click="calibrationStore.resolveAlarm('continue')"
          >
            <el-icon><CircleCheck /></el-icon>
            报警确认继续
          </el-button>
          <el-button
            type="danger"
            @click="calibrationStore.resolveAlarm('recollect')"
          >
            <el-icon><RefreshRight /></el-icon>
            报警重采
          </el-button>
        </template>
        <template v-else>
          <el-button
            type="success"
            :disabled="calibrationStore.isRunning"
            @click="calibrationStore.startCalibration()"
          >
            <el-icon><VideoPlay /></el-icon>
            开始
          </el-button>
          <el-button
            :disabled="!calibrationStore.isRunning"
            @click="calibrationStore.pauseCalibration()"
          >
            <el-icon><VideoPause /></el-icon>
            暂停
          </el-button>
          <el-button
            :disabled="sessionState !== 'paused'"
            @click="calibrationStore.resumeCalibration()"
          >
            <el-icon><RefreshRight /></el-icon>
            继续
          </el-button>
          <el-button
            type="danger"
            :disabled="sessionState === 'idle' || sessionState === 'stopped'"
            @click="calibrationStore.stopCalibration()"
          >
            <el-icon><CloseBold /></el-icon>
            停止
          </el-button>
          <div class="action-divider" />
          <el-button
            type="primary"
            :disabled="!calibrationStore.hasCollectedData"
            @click="calibrationStore.fitData()"
          >
            <el-icon><DataAnalysis /></el-icon>
            拟合
          </el-button>
          <el-button @click="calibrationStore.endCalibration()">
            <el-icon><CircleClose /></el-icon>
            结束
          </el-button>
        </template>
      </div>
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
  // Data already stored via the manual-collect API response
  console.log('Manual collection complete:', data.length, 'channels')
}
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

.workbench-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
  flex-shrink: 0;
}

.control-section {
  background: #ffffff;
  border-radius: 12px;
  border: 1px solid $slate-200;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  padding: 16px;
  font-family: $font-sans;
  transition: box-shadow 0.2s ease, border-color 0.2s ease;

  &:hover {
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.07), 0 2px 4px -2px rgba(0, 0, 0, 0.05);
  }

  .control-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 16px;
  }
}

.mode-switches {
  display: flex;
  gap: 20px;
}

.channel-select-group {
  display: flex;
  align-items: center;
  gap: 8px;

  .channel-label {
    color: $slate-500;
    font-size: 12px;
    font-weight: 500;
    letter-spacing: 0.05em;
  }
}

.switch-group {
  display: flex;
  flex-direction: column;
  gap: 6px;

  > span {
    color: $slate-500;
    font-size: 12px;
    font-weight: 500;
    letter-spacing: 0.05em;
  }
}

.progress-section {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 220px;
  flex: 1;
  max-width: 400px;

  .progress-info {
    display: flex;
    align-items: center;
    gap: 12px;
    color: $slate-600;
    font-size: 12px;
    font-weight: 500;

    .el-progress {
      flex: 1;
      min-width: 80px;
    }
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

/* 状态徽章：按 Tags 规范 */
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

.action-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  align-items: center;
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
  .mode-switches {
    gap: 16px;
  }

  .control-row {
    gap: 12px;
  }
}

@media (max-width: 900px) {
  .control-row {
    flex-direction: column;
    align-items: stretch;
  }

  .action-buttons {
    justify-content: flex-start;
  }
}
</style>
