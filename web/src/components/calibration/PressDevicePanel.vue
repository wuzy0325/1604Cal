<template>
  <div class="device-panel">
    <div class="panel-header">
      <div class="device-info">
        <el-icon class="device-icon">
          <FirstAidKit />
        </el-icon>
        <div>
          <div class="device-name">
            打压设备
          </div>
          <div class="device-type">
            压力控制器
          </div>
        </div>
      </div>
      <DeviceStatusBadge :status="deviceStatus" />
    </div>

    <div class="selection-control">
      <el-select
        v-model="selectedDeviceId"
        placeholder="选择打压设备"
        :disabled="isConnected"
        size="small"
        class="device-select"
      >
        <el-option
          v-for="dev in pressureDevices"
          :key="dev.id"
          :label="dev.name || dev.model || '未命名设备'"
          :value="dev.id"
        >
          <div class="device-option">
            <span class="device-option-name">{{ dev.name || dev.model || '未命名设备' }}</span>
            <span :class="['device-option-status', `status-${dev.status}`]">
              {{ statusLabel(dev.status) }}
            </span>
          </div>
        </el-option>
        <template #empty>
          <div class="empty-hint">
            暂无设备，请先在设备管理中添加
          </div>
        </template>
      </el-select>
      <el-button
        :type="isConnected ? 'danger' : 'primary'"
        :loading="isConnecting"
        :disabled="!selectedDeviceId && !isConnected"
        size="small"
        @click="toggleConnection"
      >
        {{ isConnected ? '断开' : '连接' }}
      </el-button>
    </div>

    <div
      v-if="isConnected"
      class="pressure-control"
    >
      <div class="current-pressure">
        <span class="label">当前压力:</span>
        <span class="value">{{ currentPressure?.toFixed(2) || '--' }}
          <span class="unit">{{ selectedUnit }}</span>
        </span>
      </div>
      <div class="unit-row">
        <span class="label">单位:</span>
        <el-select
          v-model="selectedUnit"
          size="small"
          class="unit-select"
          @change="onUnitChange"
        >
          <el-option
            v-for="u in unitOptions"
            :key="u.value"
            :label="u.label"
            :value="u.value"
          />
        </el-select>
      </div>
      <div class="pressure-actions">
        <el-input-number
          v-model="targetPressure"
          :precision="2"
          :step="1"
          size="small"
          class="target-input"
        />
        <el-button
          type="primary"
          size="small"
          @click="setPressure"
        >
          设定压力
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { FirstAidKit } from '@element-plus/icons-vue'
import DeviceStatusBadge from '@/components/common/DeviceStatusBadge.vue'
import { useMeasurementDeviceStore } from '@/stores/measurement/deviceStore'
import { multipressSetUnit, multipressSetPressure } from '@/services/apiClient'

const emit = defineEmits<{
  connect: [deviceId: string]
  disconnect: [deviceId: string]
  'set-pressure': [deviceId: string, pressure: number]
}>()

const deviceStore = useMeasurementDeviceStore()
const { pressureDevices } = storeToRefs(deviceStore)

const selectedDeviceId = ref('')
const selectedUnit = ref('kPa')
const targetPressure = ref(100)

const unitOptions = [
  { value: 'kPa', label: 'kPa (千帕)' },
  { value: 'MPa', label: 'MPa (兆帕)' },
  { value: 'bar', label: 'bar (巴)' },
  { value: 'psi', label: 'psi (磅/平方英寸)' },
  { value: 'mmHg', label: 'mmHg (毫米汞柱)' }
]

// 获取选中的打压设备
const device = computed(() =>
  deviceStore.pressureDevices.find(d => d.id === selectedDeviceId.value)
)

// 计算状态
const isConnected = computed(() => device.value?.status === 'connected')
const isConnecting = computed(() => device.value?.status === 'connecting')
const deviceStatus = computed(() => {
  if (!device.value) return 'disconnected'
  if (device.value.status === 'connected') return 'connected'
  if (device.value.status === 'connecting') return 'disconnected'
  if (device.value.status === 'error') return 'error'
  return 'disconnected'
})
const currentPressure = computed(() => device.value?.currentPressure)

// 自动选中第一个可用设备
watch(
  pressureDevices,
  (devices) => {
    if (!selectedDeviceId.value && devices.length > 0) {
      selectedDeviceId.value = devices[0].id
    }
    // 如果选中设备已不存在，清空选择
    if (selectedDeviceId.value && !devices.find(d => d.id === selectedDeviceId.value)) {
      selectedDeviceId.value = devices.length > 0 ? devices[0].id : ''
    }
  },
  { immediate: true }
)

// 设备连接后同步单位
watch(device, (dev) => {
  if (dev?.unit) {
    selectedUnit.value = dev.unit
  }
}, { immediate: true })

// 切换单位时同步到后端并更新本地 store
async function onUnitChange(unit: string) {
  if (!device.value) return
  device.value.unit = unit
  try {
    await multipressSetUnit(device.value.id, unit)
  } catch {
    // 静默失败，本地已更新
  }
}

function statusLabel(status: string): string {
  switch (status) {
    case 'connected': return '已连接'
    case 'connecting': return '连接中'
    case 'error': return '异常'
    default: return '未连接'
  }
}

const toggleConnection = async () => {
  if (!selectedDeviceId.value) return

  if (isConnected.value) {
    emit('disconnect', selectedDeviceId.value)
  } else {
    emit('connect', selectedDeviceId.value)
  }
}

const setPressure = async () => {
  if (!device.value) return
  try {
    await multipressSetPressure(device.value.id, targetPressure.value)
  } catch {
    // 静默失败
  }
  emit('set-pressure', device.value.id, targetPressure.value)
}
</script>

<style scoped lang="scss">
.device-panel {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);

  .panel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;

    .device-info {
      display: flex;
      align-items: center;
      gap: var(--spacing-sm);

      .device-icon {
        font-size: 18px;
        color: var(--status-info);
      }

      .device-name {
        color: var(--text-primary);
        font-weight: 500;
        font-size: 13px;
      }

      .device-type {
        color: var(--text-secondary);
        font-size: 11px;
      }
    }
  }

  .selection-control {
    display: flex;
    gap: var(--spacing-xs);
    align-items: center;

    .device-select {
      flex: 1;
    }
  }

  .device-option {
    display: flex;
    justify-content: space-between;
    align-items: center;

    .device-option-name {
      color: var(--text-primary);
      font-size: 13px;
    }

    .device-option-status {
      font-size: 10px;
      font-weight: 600;
      padding: 1px 5px;
      border-radius: 2px;
    }

    .status-connected {
      background: var(--status-success-bg);
      color: var(--status-success);
    }

    .status-connecting {
      background: var(--status-warning-bg);
      color: var(--status-warning);
    }

    .status-disconnected {
      background: var(--bg-secondary);
      color: var(--text-muted);
    }

    .status-error {
      background: var(--status-error-bg);
      color: var(--status-error);
    }
  }

  .empty-hint {
    padding: var(--spacing-sm);
    color: var(--text-muted);
    font-size: 12px;
    text-align: center;
  }

  .pressure-control {
    background: var(--bg-tertiary);
    border-radius: var(--radius-sm);
    padding: var(--spacing-sm);
    display: flex;
    flex-direction: column;
    gap: var(--spacing-sm);

    .current-pressure {
      display: flex;
      justify-content: space-between;
      align-items: baseline;

      .label {
        color: var(--text-secondary);
        font-size: 11px;
      }

      .value {
        color: var(--accent-primary);
        font-size: 20px;
        font-weight: 600;
        font-variant-numeric: tabular-nums;

        .unit {
          font-size: 11px;
          font-weight: 400;
          color: var(--text-secondary);
          margin-left: 2px;
        }
      }
    }

    .unit-row {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: var(--spacing-sm);

      .label {
        color: var(--text-secondary);
        font-size: 11px;
        flex-shrink: 0;
      }

      .unit-select {
        width: 130px;
      }
    }

    .pressure-actions {
      display: flex;
      align-items: center;
      gap: var(--spacing-xs);

      .target-input {
        flex: 1;
      }
    }
  }
}
</style>
