<template>
  <div class="manual-control-panel">
    <div class="control-row">
      <span class="control-label">目标压力</span>
      <el-input-number v-model="targetPressure" :min="0" :max="maxPressure" :step="1" size="small" />
    </div>

    <div class="control-row">
      <el-button type="primary" size="small" :loading="pressurizing" :disabled="!targetPressure" @click="handlePressurize">
        打压
      </el-button>
      <el-tag v-if="stabilityStatus" :type="stabilityStatus.isStable ? 'success' : 'warning'" size="small">
        {{ stabilityStatus.isStable ? '稳定' : '稳定中...' }}
      </el-tag>
    </div>

    <div v-if="stabilityStatus && !stabilityStatus.isStable" class="stability-progress">
      <el-progress :percentage="stabilityStatus.progress" :stroke-width="4" />
    </div>

    <el-button
      type="success"
      size="small"
      :disabled="!canCollect"
      @click="handleCollect"
    >
      采集数据
    </el-button>
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

<style scoped>
.manual-control-panel { display: flex; flex-direction: column; gap: 10px; padding: 8px 0; }
.control-row { display: flex; align-items: center; gap: 8px; }
.control-label { font-size: 13px; color: var(--text-secondary); min-width: 60px; }
.stability-progress { margin: 4px 0; }
</style>
