<template>
  <section class="selector-panel">
    <header class="selector-header">
      <div>
        <h3>设备选择</h3>
        <p>从统一设备台账中选择本模块使用的计量设备与打压设备。</p>
      </div>
      <button
        type="button"
        class="refresh-btn"
        @click="refreshDevices"
      >
        刷新设备列表
      </button>
    </header>

    <div class="selector-grid">
      <label>
        计量设备
        <select v-model="selectedMeasureDeviceId">
          <option value="">
            请选择计量设备
          </option>
          <option
            v-for="device in measureDevices"
            :key="device.id"
            :value="device.id"
          >
            {{ device.name || device.id }}（{{ statusLabel(device.status) }}）
          </option>
        </select>
      </label>

      <label>
        打压设备
        <select v-model="selectedPressureDeviceId">
          <option value="">
            请选择打压设备
          </option>
          <option
            v-for="device in pressureDevices"
            :key="device.id"
            :value="device.id"
          >
            {{ device.name || device.id }}（{{ statusLabel(device.status) }}）
          </option>
        </select>
      </label>
    </div>

    <p class="selection-summary">
      当前选择：计量设备 {{ selectedMeasureDeviceName }}；打压设备 {{ selectedPressureDeviceName }}
    </p>

    <p
      v-if="errorMessage"
      class="error"
    >
      {{ errorMessage }}
    </p>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'

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

function statusLabel(status: DeviceDTO['status']) {
  switch (status) {
    case 'connected':
      return '已连接'
    case 'connecting':
      return '连接中'
    case 'error':
      return '异常'
    default:
      return '未连接'
  }
}

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

<style scoped>
.selector-panel {
  background: #ffffff;
  border: 1px solid #d6e0ea;
  border-radius: 12px;
  padding: 12px;
}

.selector-header {
  align-items: flex-start;
  display: flex;
  gap: 12px;
  justify-content: space-between;
}

.selector-header h3 {
  color: #0f172a;
  margin: 0;
}

.selector-header p {
  color: #334155;
  margin: 6px 0 0;
}

.refresh-btn {
  background: #f8fafc;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  color: #0f172a;
  cursor: pointer;
  padding: 7px 10px;
}

.selector-grid {
  display: grid;
  gap: 10px;
  grid-template-columns: repeat(2, minmax(160px, 1fr));
  margin-top: 12px;
}

.selector-grid label {
  color: #334155;
  display: flex;
  flex-direction: column;
  font-size: 13px;
  gap: 4px;
}

.selector-grid select {
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  padding: 7px 8px;
}

.selection-summary {
  color: #0f172a;
  margin: 12px 0 0;
}

.error {
  color: #b91c1c;
  margin-top: 10px;
}

@media (max-width: 900px) {
  .selector-header {
    flex-direction: column;
  }

  .selector-grid {
    grid-template-columns: 1fr;
  }
}
</style>
