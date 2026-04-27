<template>
  <aside class="sidebar">
    <div class="sidebar-header">
      <div class="logo">
        <div class="logo-icon">
          <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M12 2L2 7L12 12L22 7L12 2Z" fill="currentColor" fill-opacity="0.2"/>
            <path d="M2 17L12 22L22 17" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M2 12L12 17L22 12" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>
        <div class="logo-text">
          <h1>1604系统</h1>
          <span>融合平台</span>
        </div>
      </div>
    </div>

    <nav class="sidebar-nav">
      <div
        v-for="item in menuItems"
        :key="item.path"
        class="nav-item"
        :class="{ active: isActive(item.path) }"
        @click="handleNavigate(item.path)"
      >
        <el-icon class="nav-icon">
          <component :is="item.icon" />
        </el-icon>
        <div class="nav-content">
          <span class="nav-title">{{ item.title }}</span>
          <span class="nav-desc">{{ item.description }}</span>
        </div>
      </div>
    </nav>

    <div class="sidebar-footer">
      <div class="system-status">
        <div class="status-dot online"></div>
        <span>系统运行正常</span>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { House, Tools, DataLine, SetUp, Odometer } from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()

const menuItems = [
  { path: '/', icon: House, title: '首页', description: '系统概览' },
  { path: '/device-management', icon: Tools, title: '设备管理', description: '设备台账管理' },
  { path: '/measurement', icon: DataLine, title: '计量模块', description: '计量数据采集' },
  { path: '/calibration', icon: SetUp, title: '标定工作台', description: '设备标定校准' },
  { path: '/multi-pressure', icon: Odometer, title: '多设备打压', description: '多设备并行控制' }
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
  width: $sidebar-width;
  height: 100%;
  background: $sidebar-bg;
  border-right: 1px solid $border-color;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  position: relative;
  z-index: 10;
}

.sidebar-header {
  padding: $spacing-6;
  border-bottom: 1px solid $border-color;
}

.logo {
  display: flex;
  align-items: center;
  gap: $spacing-3;

  .logo-icon {
    width: 40px;
    height: 40px;
    color: $primary-400;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba($primary-500, 0.1);
    border-radius: $radius-md;

    svg {
      width: 24px;
      height: 24px;
    }
  }

  .logo-text {
    display: flex;
    flex-direction: column;

    h1 {
      font-size: $font-size-lg;
      font-weight: $font-weight-bold;
      color: $text-primary;
      line-height: 1.2;
      margin: 0;
      letter-spacing: 0.5px;
    }

    span {
      font-size: 10px;
      color: $text-tertiary;
      letter-spacing: 0.1em;
      text-transform: uppercase;
    }
  }
}

// 导航菜单
.sidebar-nav {
  flex: 1;
  padding: $spacing-4;
  overflow-y: auto;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: $spacing-3;
  padding: $spacing-3 $spacing-4;
  margin-bottom: $spacing-1;
  border-radius: $radius-md;
  cursor: pointer;
  transition: all $transition-fast;
  position: relative;
  overflow: hidden;

  &:hover {
    background: rgba($neutral-700, 0.5);

    .nav-icon {
      color: $text-primary;
    }
    
    .nav-title {
      color: $text-primary;
    }
  }

  &.active {
    background: rgba($primary-500, 0.15);
    
    &::before {
      content: '';
      position: absolute;
      left: 0;
      top: 50%;
      transform: translateY(-50%);
      height: 20px;
      width: 3px;
      background: $primary-500;
      border-radius: 0 $radius-full $radius-full 0;
    }

    .nav-icon {
      color: $primary-400;
    }

    .nav-title {
      color: $text-primary;
      font-weight: $font-weight-medium;
    }
  }
}

.nav-icon {
  font-size: 18px;
  color: $text-tertiary;
  transition: color $transition-fast;
  flex-shrink: 0;
  display: flex;
  align-items: center;
}

.nav-content {
  display: flex;
  flex-direction: column;
  gap: 0;
  flex: 1;
  min-width: 0;
}

.nav-title {
  font-size: $font-size-base;
  color: $text-secondary;
  transition: color $transition-fast;
}

.nav-desc {
  font-size: 11px;
  color: $text-muted;
  display: none; // 默认隐藏描述，保持侧边栏简洁
}

// 侧边栏底部
.sidebar-footer {
  padding: $spacing-4 $spacing-6;
  border-top: 1px solid $border-color;
  background: rgba($neutral-900, 0.3);
}

.system-status {
  display: flex;
  align-items: center;
  gap: $spacing-2;
  font-size: $font-size-xs;
  color: $text-tertiary;

  .status-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    position: relative;

    &.online {
      background: $success-500;
      box-shadow: 0 0 6px rgba($success-500, 0.4);

      &::after {
        content: '';
        position: absolute;
        inset: -3px;
        border-radius: 50%;
        border: 1px solid $success-500;
        opacity: 0.3;
        animation: pulse 2s ease-in-out infinite;
      }
    }
  }
}

@keyframes pulse {
  0%, 100% {
    transform: scale(1);
    opacity: 0.3;
  }
  50% {
    transform: scale(1.5);
    opacity: 0;
  }
}

// 响应式适配
@media (max-width: 768px) {
  .sidebar {
    width: 100%;
    height: auto;
    flex-direction: row;
    padding: $spacing-3;
    align-items: center;
    border-right: none;
    border-bottom: 1px solid $border-color;

    .sidebar-header,
    .sidebar-footer {
      display: none;
    }
  }

  .sidebar-nav {
    display: flex;
    gap: $spacing-2;
    padding: 0;
    overflow-x: auto;
    
    &::-webkit-scrollbar {
      display: none;
    }
  }

  .nav-item {
    flex-direction: column;
    padding: $spacing-2;
    margin: 0;
    min-width: 60px;
    text-align: center;
    background: transparent;

    &:hover {
      background: rgba($neutral-700, 0.3);
    }

    &.active {
      background: rgba($primary-500, 0.1);
      
      &::before {
        display: none;
      }
    }
  }
  
  .nav-title {
    font-size: 10px;
  }
}
</style>
