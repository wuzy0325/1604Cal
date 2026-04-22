<template>
  <div class="alarm-config-panel">
    <div class="config-row">
      <span class="config-label">启用报警</span>
      <el-switch v-model="config.enabled" @change="emitChange" />
    </div>
    <div class="config-row">
      <span class="config-label">报警阈值 (%)</span>
      <el-input-number v-model="config.precisionThreshold" :min="0.01" :max="100" :step="0.1" size="small" @change="emitChange" />
    </div>
    <div class="config-row">
      <span class="config-label">报警声音</span>
      <el-switch v-model="config.soundEnabled" @change="emitChange" />
    </div>
    <div class="config-row">
      <span class="config-label">报警需确认</span>
      <el-switch v-model="config.confirmOnAlarm" @change="emitChange" />
    </div>
    <div class="config-row">
      <span class="config-label">启用通道</span>
      <el-select v-model="config.enabledChannels" multiple placeholder="全部通道" size="small" @change="emitChange">
        <el-option v-for="ch in 16" :key="ch" :label="`通道 ${ch}`" :value="ch" />
      </el-select>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, watch } from 'vue'
import type { AlarmConfigPayload } from '@/api/calibration'

const props = defineProps<{
  modelValue?: AlarmConfigPayload
}>()

const emit = defineEmits<{
  'update:modelValue': [value: AlarmConfigPayload]
  'change': [value: AlarmConfigPayload]
}>()

const config = reactive<AlarmConfigPayload>({
  enabled: true,
  precisionThreshold: 5.0,
  soundEnabled: true,
  confirmOnAlarm: true,
  enabledChannels: [],
  ...props.modelValue
})

watch(() => props.modelValue, (val) => {
  if (val) Object.assign(config, val)
}, { deep: true })

function emitChange() {
  emit('update:modelValue', { ...config })
  emit('change', { ...config })
}
</script>

<style scoped>
.alarm-config-panel { display: flex; flex-direction: column; gap: 12px; }
.config-row { display: flex; align-items: center; justify-content: space-between; }
.config-label { font-size: 13px; color: var(--text-secondary); }
</style>
