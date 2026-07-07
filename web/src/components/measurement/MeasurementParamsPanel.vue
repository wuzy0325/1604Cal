<template>
  <section class="params-card">
    <div class="params-row">
      <!-- 最小值 -->
      <div class="control-group">
        <label>最小:</label>
        <input
          v-model.number="measurementStore.measurementParams.minPressure"
          type="number"
          step="0.1"
          class="compact-input"
        >
      </div>

      <!-- 最大值 -->
      <div class="control-group">
        <label>最大:</label>
        <input
          v-model.number="measurementStore.measurementParams.maxPressure"
          type="number"
          step="0.1"
          class="compact-input"
        >
      </div>

      <!-- 测点数 -->
      <div class="control-group">
        <label>点数:</label>
        <input
          v-model.number="measurementStore.measurementParams.pointCount"
          type="number"
          min="2"
          max="11"
          class="compact-input narrow"
        >
      </div>

      <!-- 显示精度 -->
      <div class="control-group">
        <label>显示精度:</label>
        <input
          v-model.number="measurementStore.measurementParams.precision"
          type="number"
          min="0"
          max="6"
          class="compact-input narrow"
        >
      </div>

      <!-- 重复采样次数 -->
      <div class="control-group">
        <label>重复采样:</label>
        <input
          v-model.number="measurementStore.measurementParams.averageCount"
          type="number"
          min="1"
          max="10"
          class="compact-input narrow"
        >
      </div>

      <!-- 稳定时间 -->
      <div class="control-group">
        <label>稳定:</label>
        <select
          v-model.number="measurementStore.measurementParams.stableWaitS"
          class="compact-select"
        >
          <option :value="1">
            1s
          </option>
          <option :value="3">
            3s
          </option>
          <option :value="5">
            5s
          </option>
          <option :value="10">
            10s
          </option>
        </select>
      </div>

      <!-- 精度等级 -->
      <div class="control-group precision-level-group">
        <label>精度等级</label>
        <select
          v-model.number="p.precisionLevel"
          class="compact-select wide"
        >
          <option
            v-for="item in precisionLevelOptions"
            :key="item"
            :value="item"
          >
            {{ (item * 100).toFixed(2) }}%
          </option>
        </select>
      </div>

      <div class="divider" />

      <button
        type="button"
        class="generate-btn"
        :disabled="!isParamValid"
        @click="onGenerateClick"
      >
        生成压力表
      </button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ElMessage } from 'element-plus'
import { useMeasurementStore } from '@/stores/measurement'
import { fetchUnitConsistency } from '@/api/device'

const measurementStore = useMeasurementStore()

const precisionLevelOptions = [0.0001, 0.0002, 0.0005, 0.001, 0.002]

const p = computed(() => measurementStore.measurementParams)

const isParamValid = computed(() => {
  return (
    Number.isFinite(p.value.minPressure) &&
    Number.isFinite(p.value.maxPressure) &&
    p.value.maxPressure > p.value.minPressure &&
    p.value.pointCount >= 2 &&
    p.value.pointCount <= 11 &&
    p.value.averageCount >= 1
  )
})

async function onGenerateClick() {
  if (!isParamValid.value) {
    ElMessage.warning('请先填写有效的计量参数')
    return
  }
  const unitCheck = await fetchUnitConsistency().catch(() => null)
  if (unitCheck && !unitCheck.consistent) {
    ElMessage.warning('设备压力单位不一致，建议统一单位后再生成压力表')
  }
  await measurementStore.generatePoints()
}
</script>

<style scoped lang="scss">
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

.precision-level-group {
  .compact-select,
  .compact-input {
    min-width: 80px;
  }
}

.divider {
  width: 1px;
  height: 20px;
  background: $slate-200;
  margin: 0 4px;
}

.generate-btn {
  height: 32px;
  padding: 0 16px;
  background: linear-gradient(135deg, $mint, $mint-dark);
  color: #fff;
  font-size: 12px;
  font-weight: 600;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s ease;
  font-family: $font-sans;
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.25);

  &:hover:not(:disabled) {
    background: linear-gradient(135deg, $mint-light, $mint);
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.35);
    transform: translateY(-1px);
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

@media (max-width: 900px) {
  .params-row {
    align-items: flex-start;
  }
}
</style>
