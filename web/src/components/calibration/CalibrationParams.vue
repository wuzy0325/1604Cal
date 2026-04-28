<template>
  <section class="params-card">
    <div class="params-row">
      <div class="control-group">
        <label>最小:</label>
        <input
          v-model.number="calibrationStore.calibrationParams.minValue"
          type="number"
          step="0.1"
          class="compact-input"
        />
      </div>
      <div class="control-group">
        <label>最大:</label>
        <input
          v-model.number="calibrationStore.calibrationParams.maxValue"
          type="number"
          step="0.1"
          class="compact-input"
        />
      </div>
      <div class="control-group">
        <label>点数:</label>
        <input
          v-model.number="calibrationStore.calibrationParams.points"
          type="number"
          min="2"
          max="6"
          class="compact-input narrow"
        />
      </div>
      <div class="control-group">
        <label>精度:</label>
        <input
          v-model.number="calibrationStore.calibrationParams.precision"
          type="number"
          min="0"
          max="4"
          class="compact-input narrow"
        />
      </div>
      <div class="control-group">
        <label>平均:</label>
        <input
          v-model.number="calibrationStore.calibrationParams.averageCount"
          type="number"
          min="1"
          max="100"
          class="compact-input narrow"
        />
      </div>
      <div class="control-group">
        <label>稳定:</label>
        <select v-model.number="calibrationStore.calibrationParams.stableTime" class="compact-select">
          <option :value="1">1s</option>
          <option :value="3">3s</option>
          <option :value="5">5s</option>
          <option :value="10">10s</option>
        </select>
      </div>
      <div class="control-group">
        <label>精度等级</label>
        <select v-model="calibrationStore.calibrationParams.precisionLevel" class="compact-select wide">
          <option value="0.01">0.01%</option>
          <option value="0.05">0.05%</option>
          <option value="0.1">0.1%</option>
          <option value="0.2">0.2%</option>
        </select>
      </div>
      <div class="control-group">
        <label>打压</label>
        <div class="segment-control">
          <button
            type="button"
            class="segment-btn"
            :class="{ active: calibrationStore.calibrationParams.pressureMode === 'single' }"
            @click="calibrationStore.calibrationParams.pressureMode = 'single'"
          >单程</button>
          <button
            type="button"
            class="segment-btn"
            :class="{ active: calibrationStore.calibrationParams.pressureMode === 'roundTrip' }"
            @click="calibrationStore.calibrationParams.pressureMode = 'roundTrip'"
          >回程</button>
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

.params-card {
  background: #ffffff;
  border-radius: 12px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  border: 1px solid $slate-200;
  padding: 12px;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  font-family: $font-sans;
  transition: box-shadow 0.2s ease, border-color 0.2s ease;

  &:hover {
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.07), 0 2px 4px -2px rgba(0, 0, 0, 0.05);
  }
}

.params-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px 12px;
}

.control-group {
  display: flex;
  align-items: center;
  gap: 6px;

  label {
    font-size: 12px;
    color: $slate-500;
    font-weight: 500;
    letter-spacing: 0.05em;
    white-space: nowrap;
    font-family: $font-sans;
  }
}

.compact-input {
  height: 32px;
  font-size: 13px;
  border: 1px solid $slate-300;
  border-radius: 8px;
  padding: 0 8px;
  width: 56px;
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

  &.narrow {
    width: 40px;
  }

  &.wide {
    width: 80px;
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

.compact-select {
  height: 32px;
  font-size: 13px;
  border: 1px solid $slate-300;
  border-radius: 8px;
  padding: 0 24px 0 10px;
  color: $slate-700;
  background: #fff;
  outline: none;
  cursor: pointer;
  min-width: 52px;
  appearance: none;
  font-family: $font-sans;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%239ca3af' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='m6 9 6 6 6-6'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 8px center;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;

  &:focus {
    border-color: $mint;
    box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.15);
  }

  &.wide {
    min-width: 80px;
  }
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

@media (max-width: 900px) {
  .params-row {
    align-items: flex-start;
  }
}
</style>
