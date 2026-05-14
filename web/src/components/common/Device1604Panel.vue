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
import { fetchDevices, upsertDevice } from '@/api/device'

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
  if (device.value.status === 'connecting') return 'connecting'
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
  // 同步单位到设备配置，确保 CheckUnitConsistency 比较的是实际单位
  try {
    const devices = await fetchDevices()
    const dto = devices.find(d => d.id === selectedDeviceId.value)
    if (dto) {
      await upsertDevice({ ...dto, unit })
    }
  } catch (syncErr) {
    console.warn('同步计量设备单位到配置失败:', syncErr)
  }
}
</script>

<style scoped lang="scss">
/* ── 设计系统令牌 ── */
$font-sans: 'Inter', 'SF Pro Display', -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif;
$font-mono: 'JetBrains Mono', 'Fira Code', 'SF Mono', Consolas, monospace;
$mint: #10b981;
$mint-dark: #059669;
$slate-50: #f9fafb;
$slate-100: #f3f4f6;
$slate-200: #e5e7eb;
$slate-300: #d1d5db;
$slate-400: #9ca3af;
$slate-500: #6b7280;
$slate-600: #4b5563;
$slate-700: #374151;
$slate-800: #1f2937;
$green: #22c55e;
$red: #ef4444;
$blue: #3b82f6;
$amber: #f59e0b;

.device-panel {
  display: flex;
  flex-direction: column;
  gap: 12px;
  font-family: $font-sans;

  /* 头部：设备信息 + 状态徽章 */
  .panel-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;

    .device-info {
      display: flex;
      align-items: center;
      gap: 10px;

      .device-icon {
        font-size: 20px;
        color: $mint;
        background: rgba(16, 185, 129, 0.1);
        width: 36px;
        height: 36px;
        border-radius: 8px;
        display: flex;
        align-items: center;
        justify-content: center;
      }

      .device-name {
        color: $slate-800;
        font-weight: 600;
        font-size: 14px;
        line-height: 1.3;
      }

      .device-type {
        color: $slate-500;
        font-size: 12px;
        font-weight: 400;
        line-height: 1.3;
      }
    }
  }

  /* 连接控制区 */
  .selection-control {
    display: flex;
    gap: 8px;
    align-items: center;

    .device-select {
      flex: 1;
    }
  }

  /* 下拉选项 */
  .device-option {
    display: flex;
    justify-content: space-between;
    align-items: center;

    .device-option-name {
      color: $slate-700;
      font-size: 13px;
    }

    .device-option-status {
      font-size: 10px;
      font-weight: 600;
      padding: 2px 6px;
      border-radius: 4px;
      letter-spacing: 0.02em;
    }

    .status-connected {
      background: rgba(34, 197, 94, 0.12);
      border: 1px solid rgba(34, 197, 94, 0.25);
      color: #16a34a;
    }

    .status-connecting {
      background: rgba(245, 158, 11, 0.12);
      border: 1px solid rgba(245, 158, 11, 0.25);
      color: #d97706;
    }

    .status-disconnected {
      background: rgba(107, 114, 128, 0.08);
      border: 1px solid rgba(107, 114, 128, 0.18);
      color: $slate-500;
    }

    .status-error {
      background: rgba(239, 68, 68, 0.1);
      border: 1px solid rgba(239, 68, 68, 0.2);
      color: #dc2626;
    }
  }

  .empty-hint {
    padding: 12px;
    color: $slate-400;
    font-size: 12px;
    text-align: center;
    font-family: $font-sans;
  }

  /* 设备状态信息区 */
  .device-status {
    background: $slate-50;
    border-radius: 8px;
    border: 1px solid $slate-100;
    padding: 12px;
    display: flex;
    flex-direction: column;
    gap: 8px;

    .status-row {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 8px;

      .label {
        color: $slate-500;
        font-size: 12px;
        font-weight: 500;
        letter-spacing: 0.02em;
        flex-shrink: 0;
      }

      .value {
        color: $slate-700;
        font-size: 13px;
        font-weight: 600;
        font-family: $font-mono;
        text-align: right;
      }

      .unit-select {
        width: 110px;
      }
    }
  }

  /* 阀门控制按钮区 */
  .valve-control {
    display: flex;
    gap: 6px;

    .el-button {
      flex: 1;
      height: 32px;
      font-size: 12px;
      font-weight: 600;
      border-radius: 6px;
    }
  }
}
</style>
