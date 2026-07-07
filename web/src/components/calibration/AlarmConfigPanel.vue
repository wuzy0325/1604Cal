<template>
  <div class="alarm-config-panel">
    <div class="config-row">
      <span class="config-label">启用报警</span>
      <label class="inline-check">
        <input
          v-model="config.enabled"
          type="checkbox"
          @change="emitChange"
        >
      </label>
    </div>
    <div class="config-row">
      <span class="config-label">报警阈值 (%)</span>
      <input
        v-model.number="config.precisionThreshold"
        type="number"
        :min="0.01"
        :max="100"
        :step="0.1"
        class="compact-input"
        @change="emitChange"
      >
    </div>
    <div class="config-row">
      <span class="config-label">报警声音</span>
      <label class="inline-check">
        <input
          v-model="config.soundEnabled"
          type="checkbox"
          @change="emitChange"
        >
      </label>
    </div>
    <div class="config-row">
      <span class="config-label">报警需确认</span>
      <label class="inline-check">
        <input
          v-model="config.confirmOnAlarm"
          type="checkbox"
          @change="emitChange"
        >
      </label>
    </div>
    <div class="config-row">
      <span class="config-label">启用通道</span>
      <div class="channel-grid">
        <label
          v-for="ch in 16"
          :key="ch"
          class="inline-check"
        >
          <input
            :checked="config.enabledChannels.includes(ch)"
            type="checkbox"
            @change="toggleChannel(ch)"
          >
          <span>{{ ch }}</span>
        </label>
      </div>
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

function toggleChannel(ch: number) {
  const idx = config.enabledChannels.indexOf(ch)
  if (idx >= 0) {
    config.enabledChannels.splice(idx, 1)
  } else {
    config.enabledChannels.push(ch)
  }
  emitChange()
}
</script>

<style scoped lang="scss">
$font-sans: 'DM Sans', 'SF Pro Display', -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif;
$font-mono: 'JetBrains Mono', 'Fira Code', 'SF Mono', Consolas, monospace;
$mint: #10b981;
$slate-50: #f9fafb;
$slate-100: #f3f4f6;
$slate-200: #e5e7eb;
$slate-300: #d1d5db;
$slate-400: #9ca3af;
$slate-500: #6b7280;
$slate-600: #4b5563;
$slate-700: #374151;
$slate-800: #1f2937;

.alarm-config-panel {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.config-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.config-label {
  font-size: 12px;
  color: $slate-500;
  font-weight: 500;
  letter-spacing: 0.05em;
  white-space: nowrap;
  font-family: $font-sans;
}

.inline-check {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: $slate-600;
  font-size: 12px;
  cursor: pointer;
  white-space: nowrap;
  font-family: $font-sans;

  input[type="checkbox"] {
    width: 14px;
    height: 14px;
    accent-color: $mint;
    border: 1px solid $slate-300;
    border-radius: 3px;
    cursor: pointer;
  }

  &:hover span {
    color: $slate-800;
  }
}

.compact-input {
  height: 32px;
  font-size: 13px;
  border: 1px solid $slate-300;
  border-radius: 8px;
  padding: 0 8px;
  width: 60px;
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

.channel-grid {
  display: grid;
  grid-template-columns: repeat(8, auto);
  gap: 4px 8px;
}
</style>
