<template>
  <section class="workbench-section params-section">
    <div class="section-body">
      <div class="params-row">
        <div class="control-group">
          <label>最小值</label>
          <el-input-number
            v-model="calibrationStore.calibrationParams.minValue"
            :precision="2"
            :step="0.1"
            size="small"
          />
        </div>
        <div class="control-group">
          <label>最大值</label>
          <el-input-number
            v-model="calibrationStore.calibrationParams.maxValue"
            :precision="2"
            :step="0.1"
            size="small"
          />
        </div>
        <div class="control-group">
          <label>测点数</label>
          <el-input-number
            v-model="calibrationStore.calibrationParams.points"
            :min="2"
            :max="6"
            size="small"
          />
        </div>
        <div class="control-group">
          <label>精度</label>
          <el-input-number
            v-model="calibrationStore.calibrationParams.precision"
            :min="0"
            :max="4"
            size="small"
          />
        </div>
        <div class="control-group">
          <label>采样数</label>
          <el-input-number
            v-model="calibrationStore.calibrationParams.averageCount"
            :min="1"
            :max="100"
            size="small"
          />
        </div>
        <div class="control-group">
          <label>稳定时间</label>
          <el-select
            v-model="calibrationStore.calibrationParams.stableTime"
            size="small"
          >
            <el-option label="1s" :value="1" />
            <el-option label="3s" :value="3" />
            <el-option label="5s" :value="5" />
            <el-option label="10s" :value="10" />
          </el-select>
        </div>
        <div class="control-group">
          <label>精度等级</label>
          <el-select
            v-model="calibrationStore.calibrationParams.precisionLevel"
            size="small"
          >
            <el-option label="0.01%" value="0.01" />
            <el-option label="0.05%" value="0.05" />
            <el-option label="0.1%" value="0.1" />
            <el-option label="0.2%" value="0.2" />
          </el-select>
        </div>
        <div class="control-group">
          <label>打压模式</label>
          <el-radio-group
            v-model="calibrationStore.calibrationParams.pressureMode"
            size="small"
          >
            <el-radio-button value="single">单程</el-radio-button>
            <el-radio-button value="roundTrip">回程</el-radio-button>
          </el-radio-group>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useCalibrationStore } from '@/stores/calibration'

const calibrationStore = useCalibrationStore()

let debounceTimer: ReturnType<typeof setTimeout> | null = null

watch(
  () => calibrationStore.calibrationParams,
  (_, oldValue) => {
    // immediate 触发（启动时 oldValue === undefined）且已有保存的压力点时，跳过
    if (oldValue === undefined && calibrationStore.pressurePoints.length > 0) return
    if (debounceTimer) clearTimeout(debounceTimer)
    debounceTimer = setTimeout(() => {
      calibrationStore.generatePressurePoints()
    }, 500)
  },
  { deep: true, immediate: true }
)
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

.workbench-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
  flex-shrink: 0;
}

.params-section {
  background: #ffffff;
  border-radius: 12px;
  border: 1px solid $slate-200;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  padding: 12px;
  font-family: $font-sans;
  transition: box-shadow 0.2s ease, border-color 0.2s ease;

  &:hover {
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.07), 0 2px 4px -2px rgba(0, 0, 0, 0.05);
  }

  .section-body {
    display: flex;
    align-items: flex-end;
    gap: 12px;
    flex-wrap: wrap;
  }

  .params-row {
    display: flex;
    align-items: flex-end;
    gap: 12px;
    flex-wrap: wrap;
    flex: 1;
  }
}

.control-group {
  display: flex;
  flex-direction: column;
  gap: 6px;

  label {
    color: $slate-500;
    font-size: 12px;
    font-weight: 500;
    letter-spacing: 0.05em;
    white-space: nowrap;
    font-family: $font-sans;
  }

  :deep(.el-input-number) {
    width: 90px;
  }

  :deep(.el-select) {
    width: 85px;
  }

  :deep(.el-input-number__decrease),
  :deep(.el-input-number__increase) {
    border-color: $slate-200;
    color: $slate-400;
    &:hover { color: $slate-600; }
  }

  :deep(.el-input__wrapper) {
    box-shadow: 0 0 0 1px $slate-300 inset;
    &:focus-within {
      box-shadow: 0 0 0 1px $mint inset, 0 0 0 3px rgba(16, 185, 129, 0.15);
    }
  }
}

@media (max-width: 1400px) {
  .params-section .section-body {
    flex-direction: column;
    align-items: stretch;
  }
}

@media (max-width: 1200px) {
  .control-group {
    :deep(.el-input-number) {
      width: 80px;
    }
  }
}
</style>
