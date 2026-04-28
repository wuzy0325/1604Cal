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
import { useDeviceInventoryStore } from '@/stores/device/inventoryStore'
import {
  multipressSetUnit,
  multipressSetPressure
} from "@/api/multipress"

const emit = defineEmits<{
  connect: [deviceId: string]
  disconnect: [deviceId: string]
  'set-pressure': [deviceId: string, pressure: number]
}>()

const deviceStore = useDeviceInventoryStore()
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
        color: $blue;
        background: rgba(59, 130, 246, 0.1);
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

  /* 压力控制区（连接后显示） */
  .pressure-control {
    background: $slate-50;
    border-radius: 8px;
    border: 1px solid $slate-100;
    padding: 14px;
    display: flex;
    flex-direction: column;
    gap: 12px;

    .current-pressure {
      display: flex;
      justify-content: space-between;
      align-items: baseline;

      .label {
        color: $slate-500;
        font-size: 12px;
        font-weight: 500;
        letter-spacing: 0.02em;
      }

      .value {
        color: $mint;
        font-size: 24px;
        font-weight: 700;
        font-variant-numeric: tabular-nums;
        font-family: $font-mono;
        line-height: 1;

        .unit {
          font-size: 12px;
          font-weight: 500;
          color: $slate-500;
          margin-left: 4px;
          font-family: $font-sans;
        }
      }
    }

    .unit-row {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 10px;

      .label {
        color: $slate-500;
        font-size: 12px;
        font-weight: 500;
        letter-spacing: 0.02em;
        flex-shrink: 0;
      }

      .unit-select {
        width: 130px;
      }
    }

    .pressure-actions {
      display: flex;
      align-items: center;
      gap: 8px;
      padding-top: 4px;
      border-top: 1px solid $slate-200;

      .target-input {
        flex: 1;
      }

      .el-button {
        height: 32px;
        font-size: 12px;
        font-weight: 600;
        border-radius: 6px;
      }
    }
  }
}
</style>
