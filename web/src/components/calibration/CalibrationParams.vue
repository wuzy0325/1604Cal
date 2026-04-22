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
.workbench-section {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
  flex-shrink: 0;
}

.params-section {
  background: var(--bg-secondary);
  border-radius: var(--radius-md);
  padding: var(--spacing-md) var(--spacing-lg);

  .section-body {
    display: flex;
    align-items: flex-end;
    gap: var(--spacing-lg);
    flex-wrap: wrap;
  }

  .params-row {
    display: flex;
    align-items: flex-end;
    gap: var(--spacing-md);
    flex-wrap: wrap;
    flex: 1;
  }
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
    width: 90px;
  }

  :deep(.el-select) {
    width: 85px;
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
