<template>
  <aside
    class="sidebar"
    :class="{ collapsed: collapsed }"
  >
    <div
      class="sidebar-toggle"
      role="button"
      tabindex="0"
      aria-label="切换侧边栏"
      @click="emit('toggle')"
      @keydown.enter="emit('toggle')"
      @keydown.space.prevent="emit('toggle')"
    >
      <el-icon>
        <ArrowRight v-if="collapsed" />
        <ArrowLeft v-else />
      </el-icon>
    </div>
    <div
      v-show="!collapsed"
      class="sidebar-content"
    >
      <div class="sidebar-section">
        <h3 class="sidebar-title">
          <el-icon><Monitor /></el-icon>
          1604 计量设备
        </h3>
        <MeasurementDevicePanel
          @connect="handleMeasureDeviceConnect"
          @disconnect="handleMeasureDeviceDisconnect"
          @unit-change="handleMeasureUnitChange"
        />
      </div>
      <div class="sidebar-section">
        <h3 class="sidebar-title">
          <el-icon><FirstAidKit /></el-icon>
          打压设备
        </h3>
        <PressDevicePanel
          @connect="handlePressDeviceConnect"
          @disconnect="handlePressDeviceDisconnect"
          @set-pressure="handleSetPressure"
          @exhaust="handleExhaust"
          @unit-change="handlePressUnitChange"
        />
      </div>
      <div class="sidebar-section">
        <h3 class="sidebar-title">
          <el-icon><CircleCheckFilled /></el-icon>
          启动条件
        </h3>
        <div class="prerequisites-list">
          <div v-for="(item, index) in prerequisites" :key="index" class="prereq-item" :class="{ satisfied: item.satisfied, unsatisfied: !item.satisfied }">
            <el-icon v-if="item.satisfied" class="icon-satisfied"><CircleCheckFilled /></el-icon>
            <el-icon v-else class="icon-unsatisfied"><CircleCloseFilled /></el-icon>
            <span class="prereq-label">{{ item.label }}</span>
            <span class="prereq-status">{{ item.satisfied ? '已满足' : '未满足' }}</span>
          </div>
        </div>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  ArrowLeft, ArrowRight, CircleCheckFilled, CircleCloseFilled,
  Monitor, FirstAidKit
} from '@element-plus/icons-vue'
import MeasurementDevicePanel from '@/components/measurement/MeasurementDevicePanel.vue'
import PressDevicePanel from '@/components/common/PressDevicePanel.vue'
import { fetchUnitConsistency, fetchDevices, upsertDevice } from "@/api/device"
import { useMeasurementStore } from '@/stores/measurement'
import { useMeasurementDeviceStore } from '@/stores/measurement/deviceStore'
import { useDeviceStore } from '@/stores/deviceStore'

interface UnitCheckPayload {
  consistent: boolean
  conflicts: string[]
}

defineProps<{ collapsed: boolean }>()
const emit = defineEmits<{
  toggle: []
  unitCheck: [payload: UnitCheckPayload]
}>()

const measurementStore = useMeasurementStore()
const deviceStore = useMeasurementDeviceStore()
const moduleDeviceStore = useDeviceStore()
const unitConsistent = ref(true)
const unitConflicts = ref<string[]>([])

const hasConnectedPressureDevice = computed(() =>
  deviceStore.pressureDevices.some(d => d.status === 'connected')
)

const showUnitCheck = computed(() =>
  measurementStore.deviceBound && hasConnectedPressureDevice.value
)

const unitCheckSatisfied = computed(() =>
  showUnitCheck.value && unitConsistent.value
)

const emitUnitCheck = () => {
  emit('unitCheck', {
    consistent: unitConsistent.value,
    conflicts: [...unitConflicts.value]
  })
}

const checkUnitConsistency = async () => {
  try {
    const result = await fetchUnitConsistency()
    unitConsistent.value = result.consistent
    unitConflicts.value = result.conflicts ?? []
    emitUnitCheck()
  } catch {
    // 静默失败：保留当前 UI 状态，不打断流程。
  }
}

const handleMeasureDeviceConnect = async (deviceId: string) => {
  const ok = await deviceStore.connectMeasureDevice(deviceId)
  if (!ok) {
    return
  }

  moduleDeviceStore.setModuleSelection('measurement', { measureDeviceId: deviceId })

  const connectedPressure = deviceStore.pressureDevices.find(d => d.status === 'connected')
  if (connectedPressure) {
    await measurementStore.bindDevices(deviceId, connectedPressure.id)
  } else {
    unitConsistent.value = true
    unitConflicts.value = []
    emitUnitCheck()
    await measurementStore.bindMeasureDevice(deviceId)
  }

  await Promise.all([
    measurementStore.refreshDeviceInfo(),
    measurementStore.refreshValveStatus(),
    measurementStore.refreshMeasureUnit()
  ])

  if (measurementStore.measureUnit) {
    try {
      const devices = await fetchDevices()
      const dto = devices.find(d => d.id === deviceId)
      if (dto) {
        await upsertDevice({ ...dto, unit: measurementStore.measureUnit })
      }
    } catch (syncErr) {
      console.warn('同步计量设备单位到配置失败:', syncErr)
    }
  }

  // 单位同步到后端后再检查一致性，确保 deviceManager 中两个设备的单位都已是最新值
  if (connectedPressure) {
    await checkUnitConsistency()
  }
}

const handleMeasureDeviceDisconnect = async (deviceId: string) => {
  await deviceStore.disconnectMeasureDevice(deviceId)
  if (measurementStore.measureDeviceId === deviceId) {
    measurementStore.unbindMeasureDevice()
  }

  moduleDeviceStore.setModuleSelection('measurement', {
    measureDeviceId: '',
    pressureDeviceId: measurementStore.pressureDeviceId
  })

  unitConsistent.value = true
  unitConflicts.value = []
  emitUnitCheck()
}

const handlePressDeviceConnect = async (deviceId: string) => {
  const ok = await deviceStore.connectPressureDevice(deviceId)
  if (!ok) {
    return
  }

  moduleDeviceStore.setModuleSelection('measurement', { pressureDeviceId: deviceId })

  if (measurementStore.measureDeviceId) {
    await measurementStore.bindDevices(measurementStore.measureDeviceId, deviceId)
  }

  await checkUnitConsistency()
}

const handlePressDeviceDisconnect = async (deviceId: string) => {
  await deviceStore.disconnectPressureDevice(deviceId)
  if (measurementStore.pressureDeviceId === deviceId) {
    measurementStore.unbindPressureDevice()
  }

  moduleDeviceStore.setModuleSelection('measurement', {
    measureDeviceId: measurementStore.measureDeviceId,
    pressureDeviceId: ''
  })

  unitConsistent.value = true
  unitConflicts.value = []
  emitUnitCheck()
}

const handlePressUnitChange = async () => {
  await checkUnitConsistency()
}

const handleMeasureUnitChange = async () => {
  await checkUnitConsistency()
}

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const handleSetPressure = async (_deviceId: string, _pressure: number) => {}
// eslint-disable-next-line @typescript-eslint/no-unused-vars
const handleExhaust = async (_deviceId: string) => {}

const prerequisites = computed(() => [
  { label: '设备已选择', satisfied: measurementStore.deviceBound },
  { label: '打压设备已连接', satisfied: hasConnectedPressureDevice.value },
  { label: '已选择采集通道', satisfied: measurementStore.channels.length > 0 },
  { label: '单位一致', satisfied: unitCheckSatisfied.value }
])

onMounted(() => {
  emitUnitCheck()
})

defineExpose({ checkUnitConsistency })
</script>

<style scoped lang="scss">
.sidebar {
  width: 280px;
  background: #f6f7f6;
  border-right: 1px solid $slate-200;
  position: relative;
  transition: width 0.25s ease;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  &.collapsed { width: 32px; }
}
.sidebar-toggle {
  position: absolute;
  right: -12px;
  top: 50%;
  transform: translateY(-50%);
  width: 12px;
  height: 36px;
  background: #fff;
  border: 1px solid $slate-200;
  border-left: none;
  border-radius: 0 6px 6px 0;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  z-index: 10;
  .el-icon { color: $slate-400; font-size: 10px; }
  &:hover { background: $slate-50; .el-icon { color: $mint; } }
}
.sidebar-content {
  padding: 16px;
  overflow-y: auto;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 16px;

  &::-webkit-scrollbar { width: 4px; }
  &::-webkit-scrollbar-thumb { background: $slate-300; border-radius: 4px; }
  &::-webkit-scrollbar-track { background: transparent; }
}

/* Section 卡片：白色背景 + 圆角 + 阴影 + 边框 */
.sidebar-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
  background: #ffffff;
  border-radius: 12px;
  border: 1px solid $slate-200;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  padding: 16px;
  transition: box-shadow 0.2s ease, border-color 0.2s ease;

  &:hover {
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.07), 0 2px 4px -2px rgba(0, 0, 0, 0.05);
  }
}

/* Section 标题：左侧 Mint 竖线 + 文字 */
.sidebar-title {
  color: $slate-500;
  font-size: 12px;
  font-weight: 600;
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  font-family: $font-sans;
  position: relative;
  padding-left: 10px;

  &::before {
    content: '';
    position: absolute;
    left: 0;
    top: 50%;
    transform: translateY(-50%);
    width: 1px;
    height: 14px;
    background: $mint;
    border-radius: 0;
  }

  .el-icon { color: $mint; font-size: 14px; }
}

/* 启动条件 */
.prerequisites-list {
  display: flex;
  flex-direction: column;
  background: #ffffff;
  border-radius: 8px;
  overflow: hidden;

  .prereq-item {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 0;
    font-size: 13px;
    font-family: $font-sans;
    border-bottom: 1px solid $slate-100;

    &:last-child { border-bottom: none; }

    .el-icon { font-size: 16px; flex-shrink: 0; }
    .icon-satisfied { color: $green; }
    .icon-unsatisfied { color: $slate-300; }
    .prereq-label {
      flex: 1;
      color: $slate-600;
      font-weight: 500;
    }
    .prereq-status {
      font-size: 11px;
      padding: 2px 8px;
      border-radius: 4px;
      font-weight: 600;
      letter-spacing: 0.02em;
      flex-shrink: 0;
    }

    &.satisfied {
      .prereq-label { color: $slate-700; }
      .prereq-status {
        background: rgba(34, 197, 94, 0.12);
        border: 1px solid rgba(34, 197, 94, 0.25);
        color: #16a34a;
      }
    }
    &.unsatisfied {
      .prereq-label { color: $slate-400; }
      .prereq-status {
        background: rgba(239, 68, 68, 0.1);
        border: 1px solid rgba(239, 68, 68, 0.2);
        color: #dc2626;
      }
    }
  }
}

@media (max-width: 900px) {
  .sidebar { width: 100% !important; border-right: none; border-bottom: 1px solid $slate-200; .sidebar-toggle { display: none; } }
  .sidebar.collapsed { width: 100% !important; }
}
</style>
