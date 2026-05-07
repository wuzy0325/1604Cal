<template>
  <div class="manual-control-panel">
    <!-- 第一行：目标压力 + 打压按钮 + 稳定状态 -->
    <div class="control-row">
      <div class="input-group">
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

    <!-- 稳定进度条 -->
    <div v-if="stabilityStatus && !stabilityStatus.isStable" class="stability-progress">
      <div class="progress-track">
        <div class="progress-fill" :style="{ width: (stabilityStatus.progress || 0) + '%' }" />
      </div>
    </div>

    <!-- 第二行：采集按钮 -->
    <div class="control-row">
      <button
        type="button"
        class="ctrl-btn btn-collect"
        :disabled="!canCollect"
        @click="handleCollect"
      >
        采集数据
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { apiPost } from '@/api/client'
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
    await apiPost<{ status: string }>('/calibration/manual-pressurize', { targetPressure: targetPressure.value })
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
    const resp = await apiPost<{ data: number[] }>('/calibration/manual-collect')
    emit('collected', resp.data)
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

/* 手动控制面板 - 纵向排列，每行一组相关控件 */
.manual-control-panel {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 10px 14px;
  min-width: 220px;
  background: rgba(249, 250, 251, 0.8);
  border-radius: 10px;
  border: 1px solid rgba(229, 231, 235, 0.8);
}

/* 每行控件组 - 横向排列，不换行 */
.control-row {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: nowrap;
  min-height: 36px;
}

/* 输入框组 - 标签+输入框紧密排列 */
.input-group {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.control-label {
  font-size: 12px;
  color: $slate-500;
  font-weight: 500;
  letter-spacing: 0.05em;
  white-space: nowrap;
  font-family: $font-sans;
}

.compact-input {
  height: 32px;
  font-size: 13px;
  border: 1px solid $slate-300;
  border-radius: 8px;
  padding: 0 8px;
  width: 70px;
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

/* 按钮基础样式 */
.ctrl-btn {
  padding: 7px 14px;
  border-radius: 8px;
  border: none;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 600;
  transition: all 0.15s ease;
  font-family: $font-sans;
  white-space: nowrap;
  flex-shrink: 0;

  &:active {
    transform: translateY(1px);
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

/* 打压按钮 */
.btn-pump {
  background: linear-gradient(135deg, $mint, $mint-dark);
  color: #fff;
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.3);
  min-width: 64px;

  &:hover:not(:disabled) {
    background: linear-gradient(135deg, $mint-light, $mint);
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.4);
    transform: translateY(-1px);
  }
}

/* 采集按钮 */
.btn-collect {
  background: rgba(55, 65, 81, 0.08);
  color: $slate-700;
  border: 1px solid $slate-200;
  width: 100%;

  &:hover:not(:disabled) {
    background: rgba(55, 65, 81, 0.14);
    border-color: $slate-300;
  }
}

/* 状态标签 */
.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 3px 10px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  line-height: 1.5;
  letter-spacing: 0.02em;
  white-space: nowrap;
  flex-shrink: 0;
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

/* 稳定进度条 */
.stability-progress {
  margin: 2px 0;
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
