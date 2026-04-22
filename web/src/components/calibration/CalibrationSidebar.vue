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
      <!-- 1604 计量设备 -->
      <div class="sidebar-section">
        <h3 class="sidebar-title">
          <el-icon><Monitor /></el-icon>
          1604 计量设备
        </h3>
        <Device1604Panel
          @connect="calibrationStore.connectDevice1604"
          @disconnect="calibrationStore.disconnectDevice1604"
        />
      </div>

      <!-- 打压设备 -->
      <div class="sidebar-section">
        <h3 class="sidebar-title">
          <el-icon><FirstAidKit /></el-icon>
          打压设备
        </h3>
        <PressDevicePanel
          @connect="calibrationStore.connectPressDevice"
          @disconnect="calibrationStore.disconnectPressDevice"
        />
      </div>

      <!-- 通道选择 -->
      <div class="sidebar-section">
        <h3 class="sidebar-title">
          <el-icon><Grid /></el-icon>
          通道选择
          <span class="channel-count">{{ calibrationStore.selectedChannels.length }}/16</span>
        </h3>
        <ChannelMatrix
          :selected-channels="calibrationStore.selectedChannels"
          @update:selected-channels="calibrationStore.setSelectedChannels"
        />
      </div>

      <!-- 校准前置条件 -->
      <div class="sidebar-section">
        <h3 class="sidebar-title">
          <el-icon><CircleCheckFilled /></el-icon>
          启动条件
        </h3>
        <div class="prerequisites-list">
          <div
            v-for="(item, index) in prerequisites"
            :key="index"
            class="prereq-item"
            :class="{ satisfied: item.satisfied }"
          >
            <el-icon v-if="item.satisfied">
              <CircleCheckFilled />
            </el-icon>
            <el-icon v-else>
              <CircleClose />
            </el-icon>
            <span>{{ item.label }}</span>
          </div>
        </div>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  ArrowLeft,
  ArrowRight,
  Grid,
  Monitor,
  FirstAidKit,
  CircleCheckFilled,
  CircleClose
} from '@element-plus/icons-vue'

import Device1604Panel from '@/components/common/Device1604Panel.vue'
import PressDevicePanel from '@/components/common/PressDevicePanel.vue'
import ChannelMatrix from '@/components/common/ChannelMatrix.vue'
import { useCalibrationStore } from '@/stores/calibration'

defineProps<{
  collapsed: boolean
}>()

defineEmits<{
  toggle: []
}>()

const calibrationStore = useCalibrationStore()

const prerequisites = computed(() => [
  { label: '1604 设备已连接', satisfied: calibrationStore.device1604Connected },
  { label: '打压设备已连接', satisfied: calibrationStore.pressDeviceConnected },
  { label: '已选择采集通道', satisfied: calibrationStore.channelsSelected }
])
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

  &.collapsed {
    width: 32px;
  }
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

  .el-icon {
    color: var(--text-secondary);
    font-size: 10px;
  }

  &:hover {
    background: var(--bg-quaternary);

    .el-icon {
      color: var(--accent-primary);
    }
  }
}

.sidebar-content {
  padding: var(--spacing-md);
  overflow-y: auto;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xl);
}

.sidebar-section {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

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

  .el-icon {
    color: var(--accent-primary);
    font-size: 13px;
  }
}

.channel-count {
  margin-left: auto;
  color: var(--text-muted);
  font-weight: 400;
  text-transform: none;
  letter-spacing: 0;
  font-size: 11px;
}

.prerequisites-list {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  padding: var(--spacing-sm);
  display: flex;
  flex-direction: column;
  gap: 4px;

  .prereq-item {
    display: flex;
    align-items: center;
    gap: var(--spacing-xs);
    padding: 2px 0;
    color: var(--text-muted);
    font-size: 12px;

    .el-icon {
      font-size: 13px;
    }

    &.satisfied {
      color: var(--status-success);
    }
  }
}

@media (max-width: 900px) {
  .sidebar {
    width: 100% !important;
    border-right: none;
    border-bottom: 1px solid var(--border-color);

    .sidebar-toggle {
      display: none;
    }
  }

  .sidebar.collapsed {
    width: 100% !important;
  }

  .sidebar-content {
    padding: var(--spacing-md);
    gap: var(--spacing-xl);
  }
}
</style>
