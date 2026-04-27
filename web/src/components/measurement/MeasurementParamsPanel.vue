<template>
  <section class="params-panel">
    <div class="params-row">
      <div class="control-group">
        <label>最小值(kPa)</label>
        <el-input-number
          v-model.number="measurementStore.measurementParams.minPressure"
          :precision="2"
          :step="0.1"
          size="small"
        />
      </div>

      <div class="control-group">
        <label>最大值(kPa)</label>
        <el-input-number
          v-model.number="measurementStore.measurementParams.maxPressure"
          :precision="2"
          :step="0.1"
          size="small"
        />
      </div>

      <div class="control-group">
        <label>测点数</label>
        <el-input-number
          v-model.number="measurementStore.measurementParams.pointCount"
          :min="2"
          :max="11"
          size="small"
        />
      </div>

      <div class="control-group">
        <label>精度</label>
        <el-input-number
          v-model.number="measurementStore.measurementParams.precision"
          :min="0"
          :max="6"
          size="small"
        />
      </div>

      <div class="control-group">
        <label>平均次数</label>
        <el-input-number
          v-model.number="measurementStore.measurementParams.averageCount"
          :min="1"
          :max="20"
          size="small"
        />
      </div>

      <div class="control-group">
        <label>稳定时间</label>
        <el-select
          v-model.number="measurementStore.measurementParams.stableWaitS"
          size="small"
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
            v-if="!useCustomPrecision"
            v-model.number="storePrecisionLevel"
            size="small"
          >
            <el-option
              v-for="item in precisionLevelOptions"
              :key="item"
              :label="`${(item * 100).toFixed(2)}%`"
              :value="item"
            />
          </el-select>

          <el-input-number
            v-else
            v-model.number="customPrecisionValue"
            :min="0.0001"
            :max="5"
            :step="0.0001"
            :precision="4"
            size="small"
          />

          <el-checkbox v-model="useCustomPrecision">
            自定义
          </el-checkbox>
        </div>
      </div>

      <div class="flex-spacer" />

      <el-button
        type="primary"
        size="small"
        :disabled="!isParamValid"
        @click="onGenerateClick"
      >
        生成压力表
      </el-button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useMeasurementStore } from '@/stores/measurement'

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

const useCustomPrecision = ref(false)
const customPrecisionValue = ref(p.value.precisionLevel)

const storePrecisionLevel = computed({
  get: () => useCustomPrecision.value ? customPrecisionValue.value : p.value.precisionLevel,
  set: (val: number) => {
    const normalized = Number.isFinite(val) ? Math.min(5, Math.max(0.0001, val)) : p.value.precisionLevel
    p.value.precisionLevel = Number(normalized.toFixed(4))
  }
})

watch(p, () => {
  if (!useCustomPrecision.value) {
    customPrecisionValue.value = p.value.precisionLevel
  }
}, { deep: true })

watch(customPrecisionValue, (val) => {
  if (useCustomPrecision.value && Number.isFinite(val)) {
    storePrecisionLevel.value = val
  }
})

async function onGenerateClick() {
  if (!isParamValid.value) {
    ElMessage.warning('请先填写有效的计量参数')
    return
  }
  await measurementStore.generatePoints()
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
