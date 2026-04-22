<template>
  <div class="device-panel">
    <div class="panel-header">
      <div class="device-info">
        <el-icon class="device-icon">
          <Cpu />
        </el-icon>
        <div>
          <div class="device-name">
            1604设备
          </div>
          <div class="device-type">
            计量采集设备
          </div>
        </div>
      </div>
      <DeviceStatusBadge :status="deviceStatus" />
    </div>

    <div class="selection-control">
      <el-select
        v-model="selectedDeviceId"
        placeholder="选择计量设备"
        :disabled="isConnected"
        size="small"
        class="device-select"
      >
        <el-option
          v-for="dev in measureDevices"
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
      class="device-status"
    >
      <div class="status-row">
        <span class="label">设备型号:</span>
        <span class="value">{{ displayModel }}</span>
      </div>
      <div class="status-row">
        <span class="label">通道数:</span>
        <span class="value">{{ displayChannels }}</span>
      </div>
      <div class="status-row">
        <span class="label">阀门状态:</span>
        <el-tag
          :type="valveTagType"
          size="small"
        >
          {{ valveStatusLabel }}
        </el-tag>
      </div>
      <div class="status-row">
        <span class="label">单位类型:</span>
        <el-select
          v-model="selectedMeasureUnit"
          size="small"
          class="unit-select"
          :placeholder="calibrationStore.measureUnit || '选择单位'"
          @change="handleMeasureUnitChange"
        >
          <el-option
            v-for="item in measureUnitOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </div>
    </div>

    <div
      v-if="isConnected"
      class="valve-control"
    >
      <el-button
        :type="normalizedValveStatus === 'calibration' ? 'primary' : 'default'"
        size="small"
        @click="calibrationStore.setValveStatus('calibration')"
      >
        校准模式
      </el-button>
      <el-button
        :type="normalizedValveStatus === 'measurement' ? 'primary' : 'default'"
        size="small"
        @click="calibrationStore.setValveStatus('measurement')"
      >
        测量模式
      </el-button>
      <el-button
        size="small"
        @click="calibrationStore.resetDevice()"
      >
        复位
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { Cpu } from '@element-plus/icons-vue'
import DeviceStatusBadge from '@/components/common/DeviceStatusBadge.vue'
import { useDeviceInventoryStore } from '@/stores/device/inventoryStore'
import { useCalibrationStore } from '@/stores/calibration'

const emit = defineEmits<{
  connect: [deviceId: string]
  disconnect: [deviceId: string]
}>()

const deviceStore = useDeviceInventoryStore()
const calibrationStore = useCalibrationStore()
const { measureDevices } = storeToRefs(deviceStore)

const selectedDeviceId = ref('')
const selectedMeasureUnit = ref('')

const measureUnitOptions = [
  { value: 'kPa', label: 'kPa' },
  { value: 'MPa', label: 'MPa' },
  { value: 'Pa', label: 'Pa' },
  { value: 'bar', label: 'bar' },
  { value: 'mbar', label: 'mbar' },
  { value: 'psi', label: 'psi' },
  { value: 'kgf/cm2', label: 'kgf/cm²' },
  { value: 'mmHg', label: 'mmHg' },
  { value: 'atm', label: 'atm' }
]

// 获取选中的计量设备
const device = computed(() =>
  measureDevices.value.find(d => d.id === selectedDeviceId.value)
)

// 优先显示后端读取的设备信息，降级为本地 store 数据
const displayModel = computed(() =>
  calibrationStore.deviceInfo['model'] || device.value?.model || '--'
)
const displayChannels = computed(() =>
  calibrationStore.deviceInfo['channels'] || device.value?.channels || 16
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

const normalizedValveStatus = computed(() => normalizeValveStatus(calibrationStore.valveStatus))

const valveTagType = computed(() => {
  if (normalizedValveStatus.value === 'calibration') {
    return 'success'
  }
  if (normalizedValveStatus.value === 'measurement') {
    return 'info'
  }
  return 'warning'
})

const valveStatusLabel = computed(() => {
  if (normalizedValveStatus.value === 'calibration') {
    return '校准(开启)'
  }
  if (normalizedValveStatus.value === 'measurement') {
    return '测量(关闭)'
  }

  const raw = calibrationStore.valveStatus?.trim()
  return raw ? `未知(${raw})` : '--'
})

// 自动选中第一个可用设备
watch(
  measureDevices,
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

watch(
  () => calibrationStore.measureUnit,
  (unit) => {
    const safeUnit = (unit || '').trim()
    const matched = measureUnitOptions.find(item =>
      item.value.toLowerCase() === safeUnit.toLowerCase()
    )
    selectedMeasureUnit.value = matched?.value || ''
  },
  { immediate: true }
)

watch(
  isConnected,
  (connected) => {
    if (!connected) {
      selectedMeasureUnit.value = ''
    }
  },
  { immediate: true }
)

function statusLabel(status: string): string {
  switch (status) {
    case 'connected': return '已连接'
    case 'connecting': return '连接中'
    case 'error': return '异常'
    default: return '未连接'
  }
}

function normalizeValveStatus(rawValue: string): 'calibration' | 'measurement' | '' {
  let value = (rawValue || '').trim().toLowerCase()
  if (!value) {
    return ''
  }

  if (value.startsWith('a')) {
    value = value.slice(1).trim()
  }

  if (['1', 'calibration', 'calibrate', 'open', 'opened', 'on', 'c/p'].includes(value)) {
    return 'calibration'
  }
  if (['0', '2', '3', 'measurement', 'measure', 'close', 'closed', 'off', 'run'].includes(value)) {
    return 'measurement'
  }

  const firstDigit = value.match(/\d+/)?.[0]
  if (firstDigit === '1') {
    return 'calibration'
  }
  if (firstDigit === '0' || firstDigit === '2' || firstDigit === '3') {
    return 'measurement'
  }

  return ''
}

const toggleConnection = async () => {
  if (!selectedDeviceId.value) return

  if (isConnected.value) {
    emit('disconnect', selectedDeviceId.value)
  } else {
    emit('connect', selectedDeviceId.value)
  }
}

const handleMeasureUnitChange = async (unit: string) => {
  if (!unit) return
  await calibrationStore.setMeasureUnit(unit)
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
        color: var(--accent-primary);
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

  .device-status {
    background: var(--bg-tertiary);
    border-radius: var(--radius-sm);
    padding: var(--spacing-sm);
    display: flex;
    flex-direction: column;
    gap: var(--spacing-xs);

    .status-row {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: var(--spacing-xs);

      .label {
        color: var(--text-secondary);
        font-size: 11px;
      }

      .value {
        color: var(--text-primary);
        font-size: 12px;
      }

      .unit-select {
        width: 120px;
      }
    }
  }

  .valve-control {
    display: flex;
    gap: var(--spacing-xs);

    .el-button {
      flex: 1;
    }
  }
}
</style>
