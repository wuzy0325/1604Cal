<template>
  <section class="selector-panel">
    <header class="selector-header">
      <div>
        <h3>设备选择</h3>
        <p>从统一设备台账中选择本模块使用的计量设备与打压设备。</p>
      </div>
      <button
        type="button"
        class="btn btn-ghost"
        @click="refreshDevices"
      >
        <el-icon><Refresh /></el-icon>
        刷新列表
      </button>
    </header>

    <div class="selector-grid">
      <label>
        <span>计量设备</span>
        <div class="select-wrapper">
          <select v-model="selectedMeasureDeviceId">
            <option value="">
              请选择计量设备
            </option>
            <option
              v-for="device in measureDevices"
              :key="device.id"
              :value="device.id"
            >
              {{ device.name || device.id }}
            </option>
          </select>
          <el-icon class="select-icon"><ArrowDown /></el-icon>
        </div>
      </label>

      <label>
        <span>打压设备</span>
        <div class="select-wrapper">
          <select v-model="selectedPressureDeviceId">
            <option value="">
              请选择打压设备
            </option>
            <option
              v-for="device in pressureDevices"
              :key="device.id"
              :value="device.id"
            >
              {{ device.name || device.id }}
            </option>
          </select>
          <el-icon class="select-icon"><ArrowDown /></el-icon>
        </div>
      </label>
    </div>

    <div class="selection-summary">
      <div class="summary-item">
        <el-icon><Tools /></el-icon>
        <div>
          <span class="summary-label">计量设备</span>
          <span class="summary-value">{{ selectedMeasureDeviceName }}</span>
        </div>
      </div>
      <div class="summary-divider" />
      <div class="summary-item">
        <el-icon><DArrowRight /></el-icon>
        <div>
          <span class="summary-label">打压设备</span>
          <span class="summary-value">{{ selectedPressureDeviceName }}</span>
        </div>
      </div>
    </div>

    <div
      v-if="errorMessage"
      class="error-message"
    >
      <el-icon><Warning /></el-icon>
      {{ errorMessage }}
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { Refresh, ArrowDown, Tools, DArrowRight, Warning } from '@element-plus/icons-vue'

import { fetchDevices, type DeviceDTO } from '@/services/apiClient'
import { useDeviceStore, type ModuleKey } from '@/stores/deviceStore'

interface Props {
  moduleKey: ModuleKey
}

const props = defineProps<Props>()

const deviceStore = useDeviceStore()

const devices = ref<DeviceDTO[]>([])
const errorMessage = ref('')

const measureDevices = computed(() => devices.value.filter((item) => item.type === 'measure'))
const pressureDevices = computed(() => devices.value.filter((item) => item.type === 'pressure'))

const selectedMeasureDeviceId = computed({
  get: () => deviceStore.selectionByModule(props.moduleKey).measureDeviceId,
  set: (value: string) => {
    deviceStore.setModuleSelection(props.moduleKey, { measureDeviceId: value })
  }
})

const selectedPressureDeviceId = computed({
  get: () => deviceStore.selectionByModule(props.moduleKey).pressureDeviceId,
  set: (value: string) => {
    deviceStore.setModuleSelection(props.moduleKey, { pressureDeviceId: value })
  }
})

const selectedMeasureDeviceName = computed(() => {
  const selected = measureDevices.value.find((item) => item.id === selectedMeasureDeviceId.value)
  return selected?.name || selected?.id || '未选择'
})

const selectedPressureDeviceName = computed(() => {
  const selected = pressureDevices.value.find((item) => item.id === selectedPressureDeviceId.value)
  return selected?.name || selected?.id || '未选择'
})

async function refreshDevices() {
  errorMessage.value = ''
  try {
    devices.value = await fetchDevices()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '获取设备列表失败'
  }
}

watch(
  devices,
  (list) => {
    const measureExists = list.some((item) => item.type === 'measure' && item.id === selectedMeasureDeviceId.value)
    if (!measureExists) {
      selectedMeasureDeviceId.value = ''
    }

    const pressureExists = list.some((item) => item.type === 'pressure' && item.id === selectedPressureDeviceId.value)
    if (!pressureExists) {
      selectedPressureDeviceId.value = ''
    }
  },
  { deep: false }
)

onMounted(() => {
  void refreshDevices()
})
</script>

<style scoped lang="scss">
.selector-panel {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: var(--spacing-lg);
  height: 100%;
  display: flex;
  flex-direction: column;
}

.selector-header {
  align-items: flex-start;
  display: flex;
  gap: var(--spacing-md);
  justify-content: space-between;
  margin-bottom: var(--spacing-lg);
}

.selector-header h3 {
  color: var(--text-primary);
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.selector-header p {
  color: var(--text-secondary);
  margin: var(--spacing-xs) 0 0;
  font-size: 13px;
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);
  padding: var(--spacing-sm) var(--spacing-md);
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  
  .el-icon {
    font-size: 14px;
  }
}

.btn-ghost {
  background: transparent;
  border-color: var(--border-color);
  color: var(--text-secondary);
  
  &:hover {
    background: var(--bg-tertiary);
    color: var(--text-primary);
  }
}

.selector-grid {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-lg);
}

.selector-grid label {
  color: var(--text-secondary);
  display: flex;
  flex-direction: column;
  font-size: 13px;
  gap: var(--spacing-xs);
}

.select-wrapper {
  position: relative;
}

.selector-grid select {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  color: var(--text-primary);
  padding: var(--spacing-sm) var(--spacing-lg) var(--spacing-sm) var(--spacing-sm);
  width: 100%;
  appearance: none;
  cursor: pointer;
  
  &:focus {
    outline: none;
    border-color: var(--accent-primary);
  }
}

.select-icon {
  position: absolute;
  right: var(--spacing-sm);
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-muted);
  font-size: 12px;
  pointer-events: none;
}

.selection-summary {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  margin-top: auto;
}

.summary-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  
  .el-icon {
    font-size: 16px;
    color: var(--accent-primary);
  }
  
  > div {
    display: flex;
    flex-direction: column;
  }
}

.summary-label {
  color: var(--text-muted);
  font-size: 11px;
}

.summary-value {
  color: var(--text-primary);
  font-size: 14px;
  font-weight: 500;
}

.summary-divider {
  height: 1px;
  background: var(--border-color);
  margin: var(--spacing-sm) 0;
}

.error-message {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  color: var(--status-error);
  font-size: 13px;
  margin-top: var(--spacing-md);
  padding: var(--spacing-sm);
  background: rgba(239, 68, 68, 0.1);
  border-radius: var(--radius-sm);
  
  .el-icon {
    font-size: 14px;
  }
}
</style>