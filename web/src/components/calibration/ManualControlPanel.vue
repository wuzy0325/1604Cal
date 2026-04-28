<template>
  <div class="manual-control-panel">
    <div class="control-row">
      <span class="control-label">目标压力</span>
      <input
        v-model.number="targetPressure"
        type="number"
        :min="0"
        :max="maxPressure"
        :step="1"
        class="compact-input"
      />
    </div>

    <div class="control-row">
      <button
        type="button"
        class="ctrl-btn btn-pump"
        :disabled="!targetPressure || pressurizing"
        @click="handlePressurize"
      >
        {{ pressurizing ? '打压中...' : '打压' }}
      </button>
      <span
        v-if="stabilityStatus"
        :class="['status-badge', stabilityStatus.isStable ? 'status-stable' : 'status-unstable']"
      >
        {{ stabilityStatus.isStable ? '稳定' : '稳定中...' }}
      </span>
    </div>

    <div v-if="stabilityStatus && !stabilityStatus.isStable" class="stability-progress">
      <div class="progress-track">
        <div class="progress-fill" :style="{ width: (stabilityStatus.progress || 0) + '%' }" />
      </div>
    </div>

    <button
      type="button"
      class="ctrl-btn btn-collect"
      :disabled="!canCollect"
      @click="handleCollect"
    >
      采集数据
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { requestJSON } from '@/api/client'
import type { ApiResponse } from '@/types/api'
import type { StabilityEventData } from '@/composables/useCalibrationSync'

const props = defineProps<{
  maxPressure?: number
  stabilityStatus?: StabilityEventData | null
}>()

const emit = defineEmits<{
  'collected': [data: number[]]
}>()

const targetPressure = ref(0)
const pressurizing = ref(false)
const collecting = ref(false)

const canCollect = computed(() => props.stabilityStatus?.isStable && !collecting.value)

async function handlePressurize() {
  if (!targetPressure.value) return
  pressurizing.value = true
  try {
    await requestJSON<ApiResponse<{ status: string }>>('/calibration/manual-pressurize', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ targetPressure: targetPressure.value })
    })
    ElMessage.success('已开始打压')
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '打压失败')
  } finally {
    pressurizing.value = false
  }
}

async function handleCollect() {
  collecting.value = true
  try {
    const resp = await requestJSON<ApiResponse<{ data: number[] }>>('/calibration/manual-collect', {
      method: 'POST'
    })
    emit('collected', resp.data.data)
    ElMessage.success('采集完成')
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '采集失败')
  } finally {
    collecting.value = false
  }
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

.manual-control-panel {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 8px 0;
}

.control-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.control-label {
  font-size: 12px;
  color: $slate-500;
  font-weight: 500;
  letter-spacing: 0.05em;
  white-space: nowrap;
  font-family: $font-sans;
  min-width: 60px;
}

.compact-input {
  height: 32px;
  font-size: 13px;
  border: 1px solid $slate-300;
  border-radius: 8px;
  padding: 0 8px;
  width: 80px;
  text-align: center;
  color: $slate-800;
  background: #fff;
  outline: none;
  font-variant-numeric: tabular-nums;
  font-family: $font-mono;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;

  &:focus {
    border-color: $mint;
    box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.15);
  }
}

.compact-input::-webkit-inner-spin-button,
.compact-input::-webkit-outer-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

.compact-input {
  -moz-appearance: textfield;
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

.btn-pump {
  background: linear-gradient(135deg, $mint, $mint-dark);
  color: #fff;
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.3);

  &:hover:not(:disabled) {
    background: linear-gradient(135deg, $mint-light, $mint);
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.4);
    transform: translateY(-1px);
  }
}

.btn-collect {
  background: rgba(55, 65, 81, 0.08);
  color: $slate-700;
  border: 1px solid $slate-200;

  &:hover:not(:disabled) {
    background: rgba(55, 65, 81, 0.14);
    border-color: $slate-300;
  }
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

.status-stable {
  background: rgba(16, 185, 129, 0.12);
  border: 1px solid rgba(16, 185, 129, 0.25);
  color: #059669;
}

.status-unstable {
  background: rgba(245, 158, 11, 0.12);
  border: 1px solid rgba(245, 158, 11, 0.25);
  color: #d97706;
}

.stability-progress {
  margin: 4px 0;
}

.progress-track {
  width: 100%;
  height: 4px;
  background: $slate-100;
  border-radius: 999px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, $mint, $mint-light);
  transition: width 0.3s ease;
}
</style>
