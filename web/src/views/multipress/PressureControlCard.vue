<template>
  <div class="pressure-card" :class="state.status">
    <!-- 头部：设备名 + 状态 + 注销 -->
    <div class="card-head">
      <div class="head-info">
        <span class="device-name">{{ metadata?.name || state.deviceId }}</span>
        <span class="device-model">{{ metadata?.model || '' }}</span>
      </div>
      <div class="head-actions">
        <span class="status-dot" :class="state.status" />
        <span class="status-label">{{ statusLabel }}</span>
        <el-button size="small" type="danger" plain @click="$emit('unregister', state.deviceId)">
          注销
        </el-button>
      </div>
    </div>

    <!-- 压力显示 -->
    <div class="pressure-display">
      <span class="pressure-value">{{ pressureDisplay }}</span>
      <span class="pressure-unit">{{ state.unit || 'kPa' }}</span>
      <span v-if="state.status === 'pressurizing'" class="stable-indicator" :class="{ stable: state.stable }">
        {{ state.stable ? '已稳定' : '稳定中...' }}
      </span>
    </div>

    <!-- 目标压力输入 -->
    <div class="pressure-input-row">
      <el-input-number
        v-model="targetInput"
        :min="0"
        :step="1"
        :precision="2"
        size="small"
        controls-position="right"
        class="target-input"
      />
      <el-select v-model="unitSelect" size="small" class="unit-select" @change="handleUnitChange">
        <el-option label="kPa" value="kPa" />
        <el-option label="MPa" value="MPa" />
        <el-option label="bar" value="bar" />
        <el-option label="psi" value="psi" />
      </el-select>
    </div>

    <!-- 操作按钮 -->
    <div class="action-row">
      <el-button
        size="small"
        type="primary"
        :disabled="state.status === 'error'"
        @click="handleSetPressure"
      >
        开始打压
      </el-button>
      <el-button size="small" type="danger" plain @click="$emit('stop', state.deviceId)">
        停止
      </el-button>
      <el-button size="small" type="warning" plain @click="$emit('exhaust', state.deviceId)">
        排空
      </el-button>
    </div>

    <!-- 状态信息 -->
    <div v-if="state.errorMessage" class="error-bar">
      {{ state.errorMessage }}
    </div>
    <div v-if="state.status === 'exhausting'" class="info-bar">
      正在排空压力...
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import type { MultiPressDeviceState } from '@/services/apiClient'
import type { DeviceMeta } from '@/stores/multipress'

const props = defineProps<{
  state: MultiPressDeviceState
  metadata?: DeviceMeta
}>()

const emit = defineEmits<{
  'set-pressure': [deviceId: string, target: number]
  stop: [deviceId: string]
  exhaust: [deviceId: string]
  unregister: [deviceId: string]
  'set-unit': [deviceId: string, unit: string]
}>()

const targetInput = ref(props.state.targetPressure || 0)
const unitSelect = ref(props.state.unit || 'kPa')

watch(
  () => props.state.targetPressure,
  (v) => {
    targetInput.value = v || 0
  }
)

watch(
  () => props.state.unit,
  (v) => {
    if (v) unitSelect.value = v
  }
)

const pressureDisplay = computed(() => {
  const p = props.state.currentPressure
  return typeof p === 'number' ? p.toFixed(3) : '--'
})

const statusLabel = computed(() => {
  const map: Record<string, string> = {
    idle: '空闲',
    pressurizing: '打压中',
    exhausting: '排空中',
    error: '异常'
  }
  return map[props.state.status] || props.state.status
})

function handleSetPressure() {
  emit('set-pressure', props.state.deviceId, targetInput.value)
}

function handleUnitChange(unit: string) {
  emit('set-unit', props.state.deviceId, unit)
}
</script>

<style scoped lang="scss">
.pressure-card {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  padding: var(--spacing-md);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
  transition: border-color 0.15s;

  &:hover {
    border-color: var(--border-color-strong);
  }

  &.pressurizing {
    border-left: 3px solid var(--accent-primary);
  }

  &.error {
    border-left: 3px solid var(--status-error);
  }

  &.exhausting {
    border-left: 3px solid var(--status-warning);
  }
}

.card-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.head-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.device-name {
  color: var(--text-primary);
  font-size: 14px;
  font-weight: 600;
}

.device-model {
  color: var(--text-secondary);
  font-size: 11px;
}

.head-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--text-muted);

  &.idle {
    background: var(--text-secondary);
  }

  &.pressurizing {
    background: var(--accent-primary);
    box-shadow: 0 0 6px rgba(255, 215, 0, 0.5);
  }

  &.exhausting {
    background: var(--status-warning);
  }

  &.error {
    background: var(--status-error);
  }
}

.status-label {
  color: var(--text-secondary);
  font-size: 11px;
}

.pressure-display {
  display: flex;
  align-items: baseline;
  gap: 6px;
  padding: var(--spacing-sm) 0;
}

.pressure-value {
  color: var(--accent-primary);
  font-size: 28px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}

.pressure-unit {
  color: var(--text-secondary);
  font-size: 13px;
}

.stable-indicator {
  font-size: 11px;
  color: var(--text-secondary);
  margin-left: auto;

  &.stable {
    color: var(--status-success);
  }
}

.pressure-input-row {
  display: flex;
  gap: var(--spacing-xs);

  .target-input {
    flex: 1;
  }

  .unit-select {
    width: 90px;
  }
}

.action-row {
  display: flex;
  gap: var(--spacing-xs);
}

.error-bar {
  background: rgba(244, 71, 71, 0.1);
  border: 1px solid rgba(244, 71, 71, 0.3);
  border-radius: 3px;
  padding: 4px 8px;
  color: var(--status-error);
  font-size: 11px;
}

.info-bar {
  background: rgba(220, 220, 170, 0.1);
  border: 1px solid rgba(220, 220, 170, 0.3);
  border-radius: 3px;
  padding: 4px 8px;
  color: var(--status-warning);
  font-size: 11px;
}
</style>
