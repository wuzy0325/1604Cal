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
      <div
        v-if="isValvelessDevice"
        class="status-row"
      >
        <span class="label">阀门状态:</span>
        <el-tag
          type="info"
          size="small"
        >
          无阀门设备
        </el-tag>
      </div>
      <div
        v-else
        class="status-row"
      >
        <span class="label">阀门状态:</span>
        <el-tag
          :type="valveTagType"
          size="small"
        >
          {{ valveStatusLabel }}
        </el-tag>
        <span
          v-if="isConnected && normalizedValveStatus !== 'calibration'"
          class="valve-hint"
        >
          阀门需切换到「校准模式」后才能开始计量
        </span>
      </div>
      <div class="status-row">
        <span class="label">单位类型:</span>
        <el-select
          v-model="selectedMeasureUnit"
          size="small"
          class="unit-select"
          :placeholder="measurementStore.measureUnit || '选择单位'"
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

    <!-- 1603 软件校零：偏移随设备配置持久化到本地，设备重连后自动应用 -->
    <div
      v-if="isConnected && isValvelessDevice"
      class="zero-calib-control"
    >
      <el-button
        size="small"
        type="primary"
        plain
        :loading="zeroCalibPending"
        :disabled="zeroCalibPending"
        @click="handleZeroCalibrate"
      >
        校零
      </el-button>
      <span class="zero-calib-hint">记录当前各通道读数作为零点偏移，保存到本地并自动扣除</span>
    </div>

    <!-- DAQ-P-1603 无阀门协议命令，隐藏阀门切换与复位控件（驱动层阀门桩恒放行门禁） -->
    <div
      v-if="isConnected && !isValvelessDevice"
      class="valve-control"
    >
      <el-button
        :type="normalizedValveStatus === 'calibration' ? 'primary' : 'default'"
        size="small"
        :loading="valvePending"
        :disabled="valvePending"
        @click="handleValveClick('calibration')"
      >
        校准模式
      </el-button>
      <el-button
        :type="normalizedValveStatus === 'measurement' ? 'primary' : 'default'"
        size="small"
        :loading="valvePending"
        :disabled="valvePending"
        @click="handleValveClick('measurement')"
      >
        测量模式
      </el-button>
      <el-button
        size="small"
        :disabled="valvePending"
        @click="measurementStore.resetDevice()"
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
import { ElMessage } from 'element-plus'
import DeviceStatusBadge from '@/components/common/DeviceStatusBadge.vue'
import { useDeviceInventoryStore } from '@/stores/device/inventoryStore'
import { useMeasurementStore } from '@/stores/measurement'
import { fetchDevices, upsertDevice } from '@/api/device'
import { calibrateZero } from '@/api/session'
import { useValveControl } from '@/composables/useValveControl'
import {
  normalizeValveStatus as normalizeValveState,
  valveTagType as toValveTagType,
  valveStatusLabel as toValveStatusLabel,
  type ValveState,
} from '@/types/valve'

const emit = defineEmits<{
  connect: [deviceId: string]
  disconnect: [deviceId: string]
  'unit-change': [payload: { deviceId: string; unit: string }]
}>()

const deviceStore = useDeviceInventoryStore()
const measurementStore = useMeasurementStore()
const { measureDevices } = storeToRefs(deviceStore)

const selectedDeviceId = ref('')
const selectedMeasureUnit = ref('')
const zeroCalibPending = ref(false)

// 阀门切换交互（pending / 写后回读 / toast 四态反馈）统一由 composable 提供。
const { valvePending, setValve: handleValveClick } = useValveControl(
  measurementStore,
  { scenario: 'measurement' }
)

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

const device = computed(() =>
  measureDevices.value.find(d => d.id === selectedDeviceId.value)
)

// DAQ-P-1603 无阀门协议命令（DLL FFI 路径）：隐藏阀门控件。
// 型号归一化与后端 factory normalizeModel 语义一致（去空格 + 大小写不敏感）。
const isValvelessDevice = computed(() => {
  const model = (device.value?.model || '').replace(/\s+/g, '').toLowerCase()
  return model === 'daq-p-1603' || model === 'p1603'
})

const isConnected = computed(() => device.value?.status === 'connected')
const isConnecting = computed(() => device.value?.status === 'connecting')
const deviceStatus = computed(() => {
  if (!device.value) return 'disconnected'
  if (device.value.status === 'connected') return 'connected'
  if (device.value.status === 'connecting') return 'disconnected'
  if (device.value.status === 'error') return 'error'
  return 'disconnected'
})

const normalizedValveStatus = computed<ValveState>(() => normalizeValveState(measurementStore.valveStatus))

const valveTagType = computed(() => toValveTagType(normalizedValveStatus.value))

const valveStatusLabel = computed(() => {
  const raw = (measurementStore.valveStatus || '').trim()
  if (!raw) return '--'
  return toValveStatusLabel(normalizedValveStatus.value, raw)
})

watch(
  measureDevices,
  (devices) => {
    if (!selectedDeviceId.value && devices.length > 0) {
      selectedDeviceId.value = devices[0].id
    }
    if (selectedDeviceId.value && !devices.find(d => d.id === selectedDeviceId.value)) {
      selectedDeviceId.value = devices.length > 0 ? devices[0].id : ''
    }
  },
  { immediate: true }
)

watch(
  () => measurementStore.measureUnit,
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
      return
    }
    // 仅在设备已绑定到会话时刷新，避免使用未绑定的旧驱动导致 "not connected" 错误。
    // 绑定和刷新由 MeasurementSidebar.handleMeasureDeviceConnect 统一编排。
    if (!measurementStore.deviceBound) return
    void Promise.all([
      measurementStore.refreshDeviceInfo(),
      measurementStore.refreshValveStatus(),
      measurementStore.refreshMeasureUnit()
    ])
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
  await measurementStore.setMeasureUnit(unit)
  try {
    const devices = await fetchDevices()
    const dto = devices.find(d => d.id === selectedDeviceId.value)
    if (dto) {
      await upsertDevice({ ...dto, unit })
    }
  } catch (syncErr) {
    console.warn('同步计量设备单位到配置失败:', syncErr)
  }
  emit('unit-change', { deviceId: selectedDeviceId.value, unit })
}

// 1603 校零：对设备所有启用通道执行软件归零，偏移持久化到本地并自动扣除。
const handleZeroCalibrate = async () => {
  if (zeroCalibPending.value) return
  if (!selectedDeviceId.value) return
  // 从后端设备配置取启用通道（P1603 各通道带量程配置，启用通道才参与采集）。
  let channels: number[] = []
  try {
    const devices = await fetchDevices()
    const dto = devices.find(d => d.id === selectedDeviceId.value)
    channels = (dto?.channels ?? [])
      .filter(c => c.enabled)
      .map(c => c.index)
      .sort((a, b) => a - b)
  } catch {
    // 拉取配置失败时回退 1-16 全通道校零，避免阻塞操作。
    channels = Array.from({ length: 16 }, (_, i) => i + 1)
  }
  if (channels.length === 0) {
    ElMessage.warning('设备未配置启用通道，无法校零')
    return
  }
  zeroCalibPending.value = true
  try {
    const offsets = await calibrateZero(channels)
    const preview = channels
      .map((ch, i) => `CH${ch}=${offsets[i] ?? 0}`)
      .join('，')
    ElMessage.success(`校零完成：${preview}`)
  } catch (err) {
    ElMessage.error((err as Error | undefined)?.message || '校零失败，请检查设备连接')
  } finally {
    zeroCalibPending.value = false
  }
}
</script>

<style scoped lang="scss">
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

      .valve-hint {
        margin-left: auto;
        color: #d97706;
        font-size: 12px;
        font-weight: 500;
        background: rgba(217, 119, 6, 0.08);
        border-radius: 4px;
        padding: 2px 8px;
      }
    }
  }

  /* 1603 校零区 */
  .zero-calib-control {
    display: flex;
    align-items: center;
    gap: 10px;

    .el-button {
      flex-shrink: 0;
    }

    .zero-calib-hint {
      color: $slate-400;
      font-size: 12px;
      line-height: 1.4;
    }
  }

  /* 阀门控制按钮区 */
  .valve-control {
    display: flex;
    gap: 4px;

    .el-button {
      flex: 1;
      height: 28px;
      font-size: 12px;
      font-weight: 600;
      border-radius: 6px;
      padding: 0 6px;
    }
  }
}
</style>
