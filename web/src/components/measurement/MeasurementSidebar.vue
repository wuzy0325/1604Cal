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
      @click="$emit('toggle')"
      @keydown.enter="$emit('toggle')"
      @keydown.space.prevent="$emit('toggle')"
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
        />
      </div>
      <div
        v-if="showUnitCheck"
        class="sidebar-section"
      >
        <h3 class="sidebar-title">
          <el-icon><ScaleToOriginal /></el-icon>
          单位检查
        </h3>
        <div
          class="unit-check-result"
          :class="unitConsistent ? 'unit-ok' : 'unit-warn'"
        >
          <el-icon v-if="unitConsistent"><CircleCheckFilled /></el-icon>
          <el-icon v-else><Warning /></el-icon>
          <span>{{ unitConsistent ? '设备单位一致' : '设备单位不一致' }}</span>
        </div>
        <div v-if="!unitConsistent && unitConflicts.length > 0" class="unit-conflicts">
          <div v-for="(msg, idx) in unitConflicts" :key="idx" class="conflict-item">{{ msg }}</div>
        </div>
      </div>
      <div class="sidebar-section">
        <h3 class="sidebar-title">
          <el-icon><Grid /></el-icon>
          通道选择
          <span class="channel-count">{{ selectedChannels.length }}/16</span>
        </h3>
        <ChannelMatrix
          :selected-channels="selectedChannels"
          @update:selected-channels="(ch: number[]) => selectedChannels = ch"
        />
      </div>
      <div class="sidebar-section">
        <h3 class="sidebar-title">
          <el-icon><CircleCheckFilled /></el-icon>
          启动条件
        </h3>
        <div class="prerequisites-list">
          <div v-for="(item, index) in prerequisites" :key="index" class="prereq-item" :class="{ satisfied: item.satisfied }">
            <el-icon v-if="item.satisfied"><CircleCheckFilled /></el-icon>
            <el-icon v-else><CircleClose /></el-icon>
            <span>{{ item.label }}</span>
          </div>
        </div>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import {
  ArrowLeft, ArrowRight, CircleCheckFilled, CircleClose,
  Warning, Grid, Monitor, FirstAidKit, ScaleToOriginal
} from '@element-plus/icons-vue'
import MeasurementDevicePanel from '@/components/measurement/MeasurementDevicePanel.vue'
import PressDevicePanel from '@/components/common/PressDevicePanel.vue'
import ChannelMatrix from '@/components/common/ChannelMatrix.vue'
import { fetchUnitConsistency } from "@/api/device"
import { useMeasurementStore } from '@/stores/measurement'
import { useMeasurementDeviceStore } from '@/stores/measurement/deviceStore'

defineProps<{ collapsed: boolean }>()
defineEmits<{ toggle: [] }>()

const measurementStore = useMeasurementStore()
const deviceStore = useMeasurementDeviceStore()
const selectedChannels = ref<number[]>([])
const unitConsistent = ref(true)
const unitConflicts = ref<string[]>([])

const hasConnectedPressureDevice = computed(() =>
  deviceStore.pressureDevices.some(d => d.status === 'connected')
)

const showUnitCheck = computed(() =>
  measurementStore.deviceBound && hasConnectedPressureDevice.value
)

const checkUnitConsistency = async () => {
  try {
    const result = await fetchUnitConsistency()
    unitConsistent.value = result.consistent
    unitConflicts.value = result.conflicts ?? []
  } catch { /* 静默失败 */ }
}

const handleMeasureDeviceConnect = async (deviceId: string) => {
  const ok = await deviceStore.connectMeasureDevice(deviceId)
  if (!ok) {
    return
  }

  const connectedPressure = deviceStore.pressureDevices.find(d => d.status === 'connected')
  if (connectedPressure) {
    await measurementStore.bindDevices(deviceId, connectedPressure.id)
    await checkUnitConsistency()
  } else {
    await measurementStore.bindMeasureDevice(deviceId)
  }

  await Promise.all([
    measurementStore.refreshDeviceInfo(),
    measurementStore.refreshValveStatus(),
    measurementStore.refreshMeasureUnit()
  ])
}

const handleMeasureDeviceDisconnect = async (deviceId: string) => {
  await deviceStore.disconnectMeasureDevice(deviceId)
  if (measurementStore.measureDeviceId === deviceId) {
    measurementStore.unbindMeasureDevice()
  }
}

const handlePressDeviceConnect = async (deviceId: string) => {
  const ok = await deviceStore.connectPressureDevice(deviceId)
  if (!ok) {
    return
  }

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
}

const handleSetPressure = async (_deviceId: string, pressure: number) => {
  console.debug('设定压力:', pressure)
}

const prerequisites = computed(() => [
  { label: '设备已选择', satisfied: measurementStore.deviceBound },
  { label: '打压设备已连接', satisfied: hasConnectedPressureDevice.value },
  { label: '已选择采集通道', satisfied: selectedChannels.value.length > 0 },
  { label: '单位一致', satisfied: unitConsistent.value }
])

defineExpose({ checkUnitConsistency, selectedChannels })
</script>

<style scoped lang="scss">
.sidebar {
  width: 280px;
  background: var(--bg-secondary);
  border-right: 1px solid var(--border-color);
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
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-left: none;
  border-radius: 0 var(--radius-sm) var(--radius-sm) 0;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  z-index: 10;
  .el-icon { color: var(--text-secondary); font-size: 10px; }
  &:hover { background: var(--bg-quaternary); .el-icon { color: var(--accent-primary); } }
}
.sidebar-content {
  padding: var(--spacing-md);
  overflow-y: auto;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xl);
}
.sidebar-section { display: flex; flex-direction: column; gap: var(--spacing-sm); }
.sidebar-title {
  color: var(--text-secondary);
  font-size: 11px;
  font-weight: 600;
  margin: 0;
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  .el-icon { color: var(--accent-primary); font-size: 13px; }
}
.channel-count { margin-left: auto; color: var(--text-muted); font-weight: 400; text-transform: none; letter-spacing: 0; font-size: 11px; }
.unit-check-result {
  display: flex; align-items: center; gap: var(--spacing-xs); font-size: 12px;
  padding: var(--spacing-xs) var(--spacing-sm); border-radius: var(--radius-sm);
  &.unit-ok { background: var(--status-success-bg); color: var(--status-success); }
  &.unit-warn { background: var(--status-warning-bg); color: var(--status-warning); }
}
.unit-conflicts { display: flex; flex-direction: column; gap: 2px; .conflict-item { color: var(--status-error); font-size: 11px; padding: 2px 0; } }
.prerequisites-list {
  background: var(--bg-tertiary); border: 1px solid var(--border-color); border-radius: var(--radius-sm);
  padding: var(--spacing-sm); display: flex; flex-direction: column; gap: 4px;
  .prereq-item {
    display: flex; align-items: center; gap: var(--spacing-xs); padding: 2px 0;
    color: var(--text-muted); font-size: 12px;
    .el-icon { font-size: 13px; }
    &.satisfied { color: var(--status-success); }
  }
}
@media (max-width: 900px) {
  .sidebar { width: 100% !important; border-right: none; border-bottom: 1px solid var(--border-color); .sidebar-toggle { display: none; } }
  .sidebar.collapsed { width: 100% !important; }
}
</style>
