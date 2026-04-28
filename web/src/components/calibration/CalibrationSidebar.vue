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
            :class="{ satisfied: item.satisfied, unsatisfied: !item.satisfied }"
          >
            <el-icon v-if="item.satisfied" class="icon-satisfied"><CircleCheckFilled /></el-icon>
            <el-icon v-else class="icon-unsatisfied"><CircleClose /></el-icon>
            <span class="prereq-label">{{ item.label }}</span>
            <span class="prereq-status">{{ item.satisfied ? '已满足' : '未满足' }}</span>
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
  Monitor,
  FirstAidKit,
  CircleCheckFilled,
  CircleClose
} from '@element-plus/icons-vue'

import Device1604Panel from '@/components/common/Device1604Panel.vue'
import PressDevicePanel from '@/components/common/PressDevicePanel.vue'
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

.sidebar {
  width: 280px;
  background: #f6f7f6;
  border-right: 1px solid $slate-200;
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
  background: #fff;
  border: 1px solid $slate-200;
  border-left: none;
  border-radius: 0 6px 6px 0;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  z-index: 10;

  .el-icon {
    color: $slate-400;
    font-size: 10px;
  }

  &:hover {
    background: $slate-50;

    .el-icon {
      color: $mint;
    }
  }
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
    width: 3px;
    height: 14px;
    background: linear-gradient(180deg, $mint, $mint-dark);
    border-radius: 2px;
  }

  .el-icon {
    color: $mint;
    font-size: 14px;
  }
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

    .el-icon {
      font-size: 16px;
      flex-shrink: 0;
    }
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
  .sidebar {
    width: 100% !important;
    border-right: none;
    border-bottom: 1px solid $slate-200;

    .sidebar-toggle {
      display: none;
    }
  }

  .sidebar.collapsed {
    width: 100% !important;
  }

  .sidebar-content {
    padding: 16px;
    gap: 16px;
  }
}
</style>
