<template>
  <section class="params-panel">
    <div class="params-row">
      <div class="control-group">
        <label>最小值(kPa)</label>
        <el-input-number
          :model-value="params.minPressure"
          :precision="2"
          :step="0.1"
          size="small"
          @update:model-value="onMinPressureChange"
        />
      </div>

      <div class="control-group">
        <label>最大值(kPa)</label>
        <el-input-number
          :model-value="params.maxPressure"
          :precision="2"
          :step="0.1"
          size="small"
          @update:model-value="onMaxPressureChange"
        />
      </div>

      <div class="control-group">
        <label>测点数</label>
        <el-input-number
          :model-value="params.pointCount"
          :min="2"
          :max="20"
          size="small"
          @update:model-value="onPointCountChange"
        />
      </div>

      <div class="control-group">
        <label>精度</label>
        <el-input-number
          :model-value="params.precision"
          :min="0"
          :max="6"
          size="small"
          @update:model-value="onPrecisionChange"
        />
      </div>

      <div class="control-group">
        <label>平均次数</label>
        <el-input-number
          :model-value="params.averageCount"
          :min="1"
          :max="20"
          size="small"
          @update:model-value="onAverageCountChange"
        />
      </div>

      <div class="control-group">
        <label>稳定时间</label>
        <el-select
          :model-value="params.stableWaitS"
          size="small"
          @update:model-value="onStableWaitChange"
        >
          <el-option label="1s" :value="1" />
          <el-option label="3s" :value="3" />
          <el-option label="5s" :value="5" />
          <el-option label="10s" :value="10" />
        </el-select>
      </div>

      <div class="control-group precision-level-group">
        <label>精度 Level</label>
        <div class="precision-level-controls">
          <el-select
            v-if="!params.useCustomPrecision"
            :model-value="params.precisionLevel"
            size="small"
            @update:model-value="onPrecisionLevelChange"
          >
            <el-option
              v-for="item in precisionLevelOptions"
              :key="item"
              :label="`${item.toFixed(2)}%`"
              :value="item"
            />
          </el-select>

          <el-input-number
            v-else
            :model-value="params.precisionLevel"
            :min="0.0001"
            :max="5"
            :step="0.0001"
            :precision="4"
            size="small"
            @update:model-value="onPrecisionLevelChange"
          />

          <el-checkbox
            :model-value="params.useCustomPrecision"
            @update:model-value="onCustomPrecisionToggle"
          >
            自定义
          </el-checkbox>
        </div>
      </div>

      <div class="flex-spacer" />

      <div class="custom-points-toggle">
        <el-switch
          :model-value="params.useCustomPoints"
          size="small"
          @update:model-value="onCustomPointsToggle"
        />
        <span class="toggle-label">自定义点</span>
      </div>

      <div v-if="params.useCustomPoints" class="custom-points-row">
        <label>压力值列表</label>
        <el-input
          :model-value="params.customPointsText"
          size="small"
          placeholder="逗号分隔, 如: 0,2.5,5,7.5,10"
          @update:model-value="onCustomPointsTextChange"
        />
        <span class="custom-points-hint">{{ parsedCustomCount }} 个压力点</span>
      </div>

      <el-button
        type="primary"
        size="small"
        :disabled="!isParamValid"
        @click="$emit('regenerate')"
      >
        生成压力表
      </el-button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useMeasurementStore } from '@/stores/measurement'

interface MeasurementUiParams {
  minPressure: number
  maxPressure: number
  pointCount: number
  precision: number
  averageCount: number
  stableWaitS: number
  precisionLevel: number
  useCustomPrecision: boolean
  controlMode: 'auto' | 'manual'
  pressureMode: 'single' | 'roundTrip'
  useCustomPoints: boolean
  customPointsText: string
}

const defaultParams: MeasurementUiParams = {
  minPressure: 0,
  maxPressure: 10,
  pointCount: 5,
  precision: 3,
  averageCount: 3,
  stableWaitS: 3,
  precisionLevel: 0.05,
  useCustomPrecision: false,
  controlMode: 'auto',
  pressureMode: 'single',
  useCustomPoints: false,
  customPointsText: ''
}

const props = defineProps<{
  params?: MeasurementUiParams
}>()

const emit = defineEmits<{
  'update:params': [value: MeasurementUiParams]
  regenerate: []
}>()

const measurementStore = useMeasurementStore()

const params = computed<MeasurementUiParams>(() => {
  if (props.params) return props.params
  // 从 store config 构造，若无则用默认值
  const cfg = measurementStore.config
  if (cfg) {
    return {
      minPressure: cfg.minPressure ?? defaultParams.minPressure,
      maxPressure: cfg.maxPressure ?? defaultParams.maxPressure,
      pointCount: cfg.pointCount ?? defaultParams.pointCount,
      precision: cfg.precision ?? defaultParams.precision,
      averageCount: cfg.averageCount ?? defaultParams.averageCount,
      stableWaitS: cfg.stableWaitS ?? defaultParams.stableWaitS,
      precisionLevel: cfg.precisionLevel ?? defaultParams.precisionLevel,
      useCustomPrecision: cfg.useCustomPrecision ?? defaultParams.useCustomPrecision,
      controlMode: cfg.controlMode ?? defaultParams.controlMode,
      pressureMode: cfg.pressureMode ?? defaultParams.pressureMode,
      useCustomPoints: cfg.useCustomPoints ?? defaultParams.useCustomPoints,
      customPointsText: cfg.customPointsText ?? defaultParams.customPointsText
    }
  }
  return defaultParams
})

const precisionLevelOptions = [0.02, 0.05, 0.1, 0.2]

const isParamValid = computed(() => {
  if (params.value.useCustomPoints) {
    return parseCustomPoints(params.value.customPointsText).length >= 2
  }
  return (
    Number.isFinite(params.value.minPressure) &&
    Number.isFinite(params.value.maxPressure) &&
    params.value.pointCount >= 2 &&
    params.value.averageCount >= 1
  )
})

const parsedCustomCount = computed(() => {
  if (!params.value.useCustomPoints) return 0
  return parseCustomPoints(params.value.customPointsText).length
})

function toNumber(value: unknown, fallback: number): number {
  return typeof value === 'number' && Number.isFinite(value) ? value : fallback
}

function toInt(value: unknown, fallback: number): number {
  const numeric = toNumber(value, fallback)
  return Math.max(1, Math.round(numeric))
}

function clamp(value: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, value))
}

function updateField<K extends keyof MeasurementUiParams>(
  key: K,
  value: MeasurementUiParams[K]
) {
  emit('update:params', {
    ...params.value,
    [key]: value
  })
}

function onMinPressureChange(value: unknown) {
  updateField('minPressure', toNumber(value, params.value.minPressure))
}

function onMaxPressureChange(value: unknown) {
  updateField('maxPressure', toNumber(value, params.value.maxPressure))
}

function onPointCountChange(value: unknown) {
  updateField('pointCount', clamp(toInt(value, params.value.pointCount), 2, 20))
}

function onPrecisionChange(value: unknown) {
  updateField('precision', clamp(toInt(value, params.value.precision), 0, 6))
}

function onAverageCountChange(value: unknown) {
  updateField('averageCount', clamp(toInt(value, params.value.averageCount), 1, 20))
}

function onStableWaitChange(value: unknown) {
  updateField('stableWaitS', clamp(toInt(value, params.value.stableWaitS), 1, 10))
}

function onPrecisionLevelChange(value: unknown) {
  const normalized = clamp(toNumber(value, params.value.precisionLevel), 0.0001, 5)
  updateField('precisionLevel', Number(normalized.toFixed(4)))
}

function onCustomPrecisionToggle(value: unknown) {
  updateField('useCustomPrecision', value === true)
}

function onCustomPointsToggle(value: unknown) {
  updateField('useCustomPoints', value === true)
}

function onCustomPointsTextChange(value: unknown) {
  updateField('customPointsText', String(value ?? ''))
}

function parseCustomPoints(text: string): number[] {
  return text
    .split(/[,;\s]+/)
    .map(s => s.trim())
    .filter(s => s.length > 0)
    .map(s => Number(s))
    .filter(n => Number.isFinite(n) && !Number.isNaN(n))
}
</script>

<style scoped lang="scss">
.params-panel {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-sm) var(--spacing-md);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
  flex-shrink: 0;
}

.params-row {
  display: flex;
  align-items: flex-end;
  flex-wrap: wrap;
  gap: var(--spacing-sm) var(--spacing-md);
}

.control-group {
  display: flex;
  flex-direction: column;
  gap: 4px;

  label {
    color: var(--text-muted);
    font-size: 11px;
    font-weight: 500;
    white-space: nowrap;
  }

  :deep(.el-input-number) {
    width: 96px;
  }

  :deep(.el-select) {
    width: 96px;
  }
}

.precision-level-group {
  :deep(.el-input-number) {
    width: 106px;
  }

  :deep(.el-select) {
    width: 106px;
  }
}

.precision-level-controls {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);

  :deep(.el-checkbox) {
    margin-right: 0;
  }
}

.flex-spacer {
  flex: 1;
}

.custom-points-toggle {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  margin-left: auto;

  .toggle-label {
    color: var(--text-muted);
    font-size: 11px;
    white-space: nowrap;
  }
}

.custom-points-row {
  display: flex;
  align-items: flex-end;
  gap: var(--spacing-xs);
  width: 100%;

  label {
    color: var(--text-muted);
    font-size: 11px;
    font-weight: 500;
    white-space: nowrap;
  }

  :deep(.el-input) {
    width: 280px;
  }

  .custom-points-hint {
    color: var(--text-secondary);
    font-size: 12px;
    white-space: nowrap;
  }
}

@media (max-width: 900px) {
  .params-panel {
    padding: var(--spacing-md);
  }

  .params-row {
    align-items: flex-start;
  }

  .flex-spacer {
    display: none;
  }
}
</style>
