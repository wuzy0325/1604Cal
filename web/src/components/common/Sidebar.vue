<template>
  <aside class="sidebar" :class="{ collapsed }">
    <div class="sidebar-header">
      <div class="logo-icon" title="1604系统">
        <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M12 2L2 7L12 12L22 7L12 2Z" fill="currentColor" fill-opacity="0.2"/>
          <path d="M2 17L12 22L22 17" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          <path d="M2 12L12 17L22 12" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
      </div>
      <transition name="fade-text">
        <span v-if="!collapsed" class="brand-label">Cal1604</span>
      </transition>
    </div>

    <nav class="sidebar-nav">
      <div
        v-for="item in menuItems"
        :key="item.path"
        class="nav-item"
        :class="{ active: isActive(item.path) }"
        :title="collapsed ? item.title : undefined"
        @click="handleNavigate(item.path)"
      >
        <el-icon class="nav-icon">
          <component :is="item.icon" />
        </el-icon>
        <transition name="fade-text">
          <span v-if="!collapsed" class="nav-label">{{ item.title }}</span>
        </transition>
        <span v-if="!collapsed" class="nav-indicator" />
      </div>
    </nav>

    <div class="sidebar-footer">
      <div class="nav-item footer-item" :title="collapsed ? '返回首页' : undefined" @click="handleNavigate('/')">
        <el-icon class="nav-icon"><House /></el-icon>
        <transition name="fade-text">
          <span v-if="!collapsed" class="nav-label">首页</span>
        </transition>
      </div>
      <button class="collapse-btn" :title="collapsed ? '展开菜单' : '收起菜单'" @click="collapsed = !collapsed">
        <el-icon class="collapse-icon" :class="{ rotated: !collapsed }"><DArrowRight /></el-icon>
      </button>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  House, Tools, DataLine, SetUp, Odometer, DArrowRight
} from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()
const collapsed = ref(false)

const menuItems = [
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
$font-sans: 'Inter', 'SF Pro Display', -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif;
$mint: #10b981;
$mint-light: #34d399;
$slate-700: #374151;
$slate-800: #1f2937;
$slate-900: #111827;

$sidebar-expanded: 180px;
$sidebar-collapsed: 52px;
$transition-sidebar: 250ms cubic-bezier(0.4, 0, 0.2, 1);

.sidebar {
  width: $sidebar-expanded;
  height: 100%;
  background: linear-gradient(180deg, $slate-800 0%, $slate-900 100%);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  position: relative;
  z-index: 20;
  overflow: hidden;
  transition: width $transition-sidebar;
  font-family: $font-sans;

  &.collapsed {
    width: $sidebar-collapsed;
    align-items: center;
  }
}

/* ── 头部 ── */
.sidebar-header {
  padding: 14px 16px 12px;
  display: flex;
  align-items: center;
  gap: 10px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  width: 100%;
  flex-shrink: 0;
  min-height: 56px;
  box-sizing: border-box;

  .collapsed & {
    justify-content: center;
    padding: 14px 0 12px;
  }
}

.logo-icon {
  width: 28px;
  height: 28px;
  color: $mint-light;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(16, 185, 129, 0.12);
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.2s ease;
  flex-shrink: 0;

  &:hover {
    background: rgba(16, 185, 129, 0.2);
  }

  svg {
    width: 18px;
    height: 18px;
  }
}

.brand-label {
  font-size: 14px;
  font-weight: 700;
  color: #fff;
  letter-spacing: -0.01em;
  white-space: nowrap;
}

/* ── 导航列表 ── */
.sidebar-nav {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 10px 8px;
  overflow-y: auto;
  overflow-x: hidden;
  scrollbar-width: none;

  &::-webkit-scrollbar { display: none; }

  .collapsed & {
    padding: 10px 0;
    align-items: center;
  }
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  height: 36px;
  padding: 0 10px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
  color: rgba(255, 255, 255, 0.35);
  white-space: nowrap;

  .collapsed & {
    justify-content: center;
    padding: 0;
    width: 36px;
  }

  &:hover {
    background: rgba(255, 255, 255, 0.08);
    color: rgba(255, 255, 255, 0.7);
  }

  &.active {
    color: #fff;
    background: rgba(16, 185, 129, 0.15);

    .nav-indicator {
      opacity: 1;
    }
  }
}

.nav-icon {
  font-size: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.nav-label {
  font-size: 13px;
  font-weight: 500;
  letter-spacing: 0.01em;
}

.nav-indicator {
  position: absolute;
  left: -9px;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 16px;
  background: $mint;
  border-radius: 0 2px 2px 0;
  opacity: 0;
  transition: opacity 0.2s ease;
  box-shadow: 0 0 6px rgba(16, 185, 129, 0.4);
}

/* ── 底部 ── */
.sidebar-footer {
  padding: 8px;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex-shrink: 0;
  box-sizing: border-box;

  .collapsed & {
    align-items: center;
    padding: 8px 0;
  }
}

.footer-item {
  color: rgba(255, 255, 255, 0.2);

  &:hover {
    color: rgba(255, 255, 255, 0.5);
  }
}

.collapse-btn {
  width: 100%;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: none;
  border: none;
  color: rgba(255, 255, 255, 0.15);
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.2s ease;

  .collapsed & {
    width: 36px;
  }

  &:hover {
    color: rgba(255, 255, 255, 0.4);
    background: rgba(255, 255, 255, 0.06);
  }
}

.collapse-icon {
  font-size: 14px;
  transition: transform 0.25s ease;

  &.rotated {
    transform: rotate(180deg);
  }
}

/* ── 文字淡入淡出 ── */
.fade-text-enter-active { transition: opacity 0.2s ease 0.05s; }
.fade-text-leave-active { transition: opacity 0.1s ease; }
.fade-text-enter-from,
.fade-text-leave-to { opacity: 0; }

/* ── 响应式 ── */
@media (max-width: 768px) {
  .sidebar {
    width: 100%;
    height: auto;
    flex-direction: row;
    padding: 6px 12px;
    align-items: center;
  }

  .sidebar-header {
    padding: 0;
    margin-right: 10px;
    border-bottom: none;
    width: auto;
    min-height: unset;
  }

  .brand-label { display: none; }

  .sidebar-nav {
    flex-direction: row;
    gap: 2px;
    padding: 0;
    flex: 1;
    justify-content: flex-start;
  }

  .nav-item {
    width: 36px;
    justify-content: center;
    padding: 0;
  }

  .nav-label { display: none; }
  .nav-indicator { display: none; }

  .sidebar-footer { display: none; }
  .collapse-btn { display: none; }
}
</style>
