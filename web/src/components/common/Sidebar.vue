<template>
  <aside class="sidebar">
    <!-- Logo: 仅图标 -->
    <div class="sidebar-header">
      <div class="logo-icon" title="1604系统">
        <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M12 2L2 7L12 12L22 7L12 2Z" fill="currentColor" fill-opacity="0.2"/>
          <path d="M2 17L12 22L22 17" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          <path d="M2 12L12 17L22 12" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
      </div>
    </div>

    <!-- 主导航 -->
    <nav class="sidebar-nav">
      <div
        v-for="item in menuItems"
        :key="item.path"
        class="nav-item"
        :class="{ active: isActive(item.path) }"
        :data-tooltip="item.title"
        @click="handleNavigate(item.path)"
      >
        <el-icon class="nav-icon">
          <component :is="item.icon" />
        </el-icon>
      </div>
    </nav>

  </aside>
</template>

<script setup lang="ts">
import { useRouter, useRoute } from 'vue-router'
import {
  House, Tools, DataLine, SetUp, Odometer
} from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()

const menuItems = [
  { path: '/', icon: House, title: '首页' },
  { path: '/device-management', icon: Tools, title: '设备管理' },
  { path: '/measurement', icon: DataLine, title: '计量模块' },
  { path: '/calibration', icon: SetUp, title: '标定工作台' },
  { path: '/multi-pressure', icon: Odometer, title: '多设备打压' }
]

function isActive(path: string): boolean {
  return route.path === path
}

function handleNavigate(path: string): void {
  router.push(path)
}

</script>

<style scoped lang="scss">
.sidebar {
  width: 56px;
  height: 100%;
  background: #fff;
  border-right: 1px solid #e5e7eb;
  display: flex;
  flex-direction: column;
  align-items: center;
  flex-shrink: 0;
  position: relative;
  z-index: 10;
  overflow: hidden;
}

.sidebar-header {
  padding: 16px 0;
  display: flex;
  justify-content: center;
}

.logo-icon {
  width: 32px;
  height: 32px;
  color: #10b981;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #ecfdf5;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;

  &:hover {
    background: #d1fae5;
  }

  svg {
    width: 20px;
    height: 20px;
  }
}

.sidebar-nav {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 8px 0;
  overflow-y: auto;
  overflow-x: hidden;
  scrollbar-width: none;

  &::-webkit-scrollbar {
    display: none;
  }
}

.nav-item {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
  color: #9ca3af;

  &:hover {
    background: #f3f4f6;
    color: #4b5563;

    &::after {
      opacity: 1;
    }
  }

  &.active {
    background: #ecfdf5;
    color: #10b981;

    &::before {
      content: '';
      position: absolute;
      left: -10px;
      top: 50%;
      transform: translateY(-50%);
      width: 3px;
      height: 20px;
      background: #10b981;
      border-radius: 0 3px 3px 0;
    }
  }

  // Tooltip
  &::after {
    content: attr(data-tooltip);
    position: absolute;
    left: calc(100% + 10px);
    top: 50%;
    transform: translateY(-50%);
    background: #374151;
    color: #fff;
    padding: 4px 8px;
    border-radius: 4px;
    font-size: 11px;
    white-space: nowrap;
    opacity: 0;
    pointer-events: none;
    transition: opacity 0.2s;
    z-index: 50;
  }
}

.nav-icon {
  font-size: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
}

// 响应式适配
@media (max-width: 768px) {
  .sidebar {
    width: 100%;
    height: auto;
    flex-direction: row;
    padding: 8px 12px;
    align-items: center;
    border-right: none;
    border-bottom: 1px solid #e5e7eb;
  }

  .sidebar-header {
    padding: 0;
    margin-right: 12px;
  }

  .logo-icon {
    width: 28px;
    height: 28px;

    svg {
      width: 18px;
      height: 18px;
    }
  }

  .sidebar-nav {
    flex-direction: row;
    gap: 4px;
    padding: 0;
    flex: 1;
    justify-content: flex-start;
    overflow-x: auto;
  }

  .nav-item {
    width: 36px;
    height: 36px;

    &.active::before {
      display: none;
    }

    &::after {
      display: none;
    }
  }
}
</style>
