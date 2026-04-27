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
.workbench-section {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
  flex-shrink: 0;
}

.control-section {
  background: var(--bg-tertiary);
  border-radius: var(--radius-md);
  padding: var(--spacing-sm) var(--spacing-md);

  .control-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: var(--spacing-md);
  }
}

.mode-switches {
  display: flex;
  gap: var(--spacing-lg);
}

.channel-select-group {
  display: flex;
  align-items: center;
  gap: 6px;

  .channel-label {
    color: var(--text-muted);
    font-size: 11px;
    font-weight: 500;
  }
}

.switch-group {
  display: flex;
  flex-direction: column;
  gap: 4px;

  > span {
    color: var(--text-muted);
    font-size: 11px;
    font-weight: 500;
  }
}

.progress-section {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 220px;
  flex: 1;
  max-width: 320px;

  .progress-info {
    display: flex;
    align-items: center;
    gap: var(--spacing-sm);
    color: var(--text-secondary);
    font-size: 13px;

    .el-progress {
      flex: 1;
      min-width: 80px;
    }
  }

  .stable-status {
    display: flex;
    gap: var(--spacing-md);
    color: var(--text-muted);
    font-size: 11px;

    .session-state {
      color: var(--accent-primary);
      font-weight: 500;
    }
  }
}

.session-status-inline {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.valve-warning {
  color: var(--status-warning);
  font-size: 11px;
  font-weight: 500;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 4px 12px;
  border-radius: var(--radius-sm);
  font-size: 12px;
  font-weight: 600;
}

.status-idle { background: var(--bg-quaternary); color: var(--text-secondary); }
.status-ready { background: var(--status-info-bg); color: var(--status-info); }

.status-pressurizing,
.status-stabilizing,
.status-collecting,
.status-point_done,
.status-fitting,
.status-await_manual_collect,
.status-await_alarm_resolution,
.status-recovering { background: var(--status-success-bg); color: var(--status-success); }

.status-paused { background: var(--status-warning-bg); color: var(--status-warning); }
.status-completed { background: var(--status-success-bg); color: var(--status-success); }
.status-stopped { background: var(--bg-quaternary); color: var(--text-secondary); }
.status-error { background: var(--status-error-bg); color: var(--status-error); }

.action-buttons {
  display: flex;
  gap: var(--spacing-xs);
  flex-wrap: wrap;
  align-items: center;
}

.action-divider {
  width: 1px;
  height: 20px;
  background: var(--border-color-strong);
  margin: 0 var(--spacing-xs);
}

.manual-controls {
  display: flex;
  flex-direction: column;
  min-width: 240px;
}

@media (max-width: 1200px) {
  .mode-switches {
    gap: var(--spacing-md);
  }

  .control-row {
    gap: var(--spacing-sm);
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
